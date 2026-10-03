package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:8080", "адрес HTTP-сервера")
	audioDir := flag.String("audio", "../audio", "папка с записями (внутри sounds/ и words/)")
	staticDir := flag.String("static", "", "собранный интерфейс (web/dist); пусто — не раздавать")
	admin := flag.Bool("admin", false, "разрешить админку: генерацию вариантов и выбор записи (macOS)")
	candDir := flag.String("candidates", "../candidates", "папка для вариантов озвучки")
	ttsScript := flag.String("tts", "../tools/tts.py", "скрипт синтезатора Silero")
	missing := flag.Bool("missing", false, "показать, каких записей не хватает, и выйти")
	pregen := flag.Bool("voices", false, "сгенерировать варианты всеми голосами там, где их нет, и выйти")
	pregenAll := flag.Bool("voices-all", false, "сгенерировать все варианты заново и выйти")
	flag.Parse()

	page, err := loadPage()
	if err != nil {
		log.Fatal(err)
	}
	lib := &Library{dir: *audioDir}
	cands := &Candidates{dir: *candDir}
	tts := &TTS{script: *ttsScript}
	defer tts.Close()

	switch {
	case *missing:
		if err := printMissing(page, lib); err != nil {
			log.Fatal(err)
		}
		return
	case *pregen || *pregenAll:
		if err := pregenerate(page, cands, tts, *pregenAll); err != nil {
			tts.Close()
			log.Fatal(err)
		}
		return
	}

	s := &server{page: page, lib: lib, cands: cands, tts: tts, admin: *admin}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/page", s.handlePage)
	mux.HandleFunc("GET /api/audio/{kind}/{key}", s.handleAudio)
	if *admin {
		mux.HandleFunc("GET /api/admin/voices", s.handleVoices)
		mux.HandleFunc("GET /api/admin/clips/{kind}/{key}", s.withClip(s.handleClip))
		mux.HandleFunc("POST /api/admin/clips/{kind}/{key}/candidates/{voice}", s.withClip(s.handleGenerate))
		mux.HandleFunc("POST /api/admin/clips/{kind}/{key}/choose", s.withClip(s.handleChoose))
		mux.HandleFunc("GET /api/admin/candidates/{kind}/{key}/{voice}", s.withClip(s.handleCandidate))
	}
	if *staticDir != "" {
		mux.Handle("GET /", http.FileServer(http.Dir(*staticDir)))
	}

	mode := ""
	if *admin {
		mode = ", админка включена"
	}
	log.Printf("domashka слушает http://%s (записи: %s%s)", *addr, *audioDir, mode)
	log.Fatal(http.ListenAndServe(*addr, mux))
}

type server struct {
	page  *Page
	lib   *Library
	cands *Candidates
	tts   *TTS
	admin bool
}

type pageResponse struct {
	Page *Page `json:"page"`
	// Recordings: «s:ш» / «w:дом» → адрес записи. Чего нет в словаре — ещё не записано.
	Recordings map[string]string `json:"recordings"`
	Total      int               `json:"total"`

	// Только при -admin: можно ли включать админку, порядок записей и чем они озвучены.
	Admin   bool              `json:"admin,omitempty"`
	Clips   []string          `json:"clips,omitempty"`
	Choices map[string]Choice `json:"choices,omitempty"`
}

func (s *server) handlePage(w http.ResponseWriter, r *http.Request) {
	have, err := s.lib.Scan()
	if err != nil {
		httpError(w, err)
		return
	}
	clips := s.page.Clips()
	resp := pageResponse{Page: s.page, Recordings: map[string]string{}, Total: len(clips)}
	for _, c := range clips {
		if rec, ok := have[Clip{Kind: c.Kind, Key: normalizeKey(c.Key)}]; ok {
			resp.Recordings[c.ID()] = recordingURL(c, rec)
		}
	}
	if s.admin {
		resp.Admin = true
		for _, c := range clips {
			resp.Clips = append(resp.Clips, c.ID())
		}
		if resp.Choices, err = s.lib.Choices(); err != nil {
			httpError(w, err)
			return
		}
	}
	writeJSON(w, resp)
}

func recordingURL(c Clip, rec Recording) string {
	return fmt.Sprintf("/api/audio/%s/%s?v=%d", c.Kind, url.PathEscape(c.Key), rec.Version)
}

func (s *server) handleAudio(w http.ResponseWriter, r *http.Request) {
	kind := Kind(r.PathValue("kind"))
	if kind != KindSound && kind != KindWord {
		http.NotFound(w, r)
		return
	}
	rec, ok, err := s.lib.Lookup(Clip{Kind: kind, Key: r.PathValue("key")})
	if err != nil {
		httpError(w, err)
		return
	}
	if !ok {
		http.NotFound(w, r)
		return
	}
	// В адресе есть ?v=<время изменения>, так что перезаписанный файл получит новый адрес.
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	http.ServeFile(w, r, rec.Path)
}

// ── Админка ───────────────────────────────────────────────

func (s *server) handleVoices(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, voices)
}

// withClip пускает дальше только записи, которые есть на странице: пути к файлам
// строятся из ключа страницы, а не из адреса.
func (s *server) withClip(h func(http.ResponseWriter, *http.Request, Clip)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, ok := s.page.FindClip(Kind(r.PathValue("kind")), r.PathValue("key"))
		if !ok {
			http.NotFound(w, r)
			return
		}
		h(w, r, c)
	}
}

type clipResponse struct {
	Text       string      `json:"text"` // каким текстом делать варианты
	Candidates []Candidate `json:"candidates"`
	Chosen     *Choice     `json:"chosen,omitempty"`
}

func (s *server) clipState(c Clip) (clipResponse, error) {
	cands, text, err := s.cands.List(c)
	if err != nil {
		return clipResponse{}, err
	}
	choices, err := s.lib.Choices()
	if err != nil {
		return clipResponse{}, err
	}
	resp := clipResponse{Text: text, Candidates: cands}
	if ch, ok := choices[c.ID()]; ok {
		resp.Chosen = &ch
	}
	if resp.Text == "" {
		resp.Text = s.page.Spoken(c)
	}
	return resp, nil
}

func (s *server) handleClip(w http.ResponseWriter, r *http.Request, c Clip) {
	resp, err := s.clipState(c)
	if err != nil {
		httpError(w, err)
		return
	}
	writeJSON(w, resp)
}

func (s *server) handleGenerate(w http.ResponseWriter, r *http.Request, c Clip) {
	v, ok := voiceByID(r.PathValue("voice"))
	if !ok {
		http.NotFound(w, r)
		return
	}
	var body struct{ Text string }
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&body); err != nil || body.Text == "" {
		http.Error(w, "нужен text", http.StatusBadRequest)
		return
	}
	cand, err := s.cands.Generate(s.tts, c, v, body.Text)
	if err != nil {
		log.Printf("%s %s: %v", c.ID(), v.ID, err)
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	writeJSON(w, cand)
}

func (s *server) handleCandidate(w http.ResponseWriter, r *http.Request, c Clip) {
	if _, ok := voiceByID(r.PathValue("voice")); !ok {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	http.ServeFile(w, r, s.cands.Path(c, r.PathValue("voice")))
}

func (s *server) handleChoose(w http.ResponseWriter, r *http.Request, c Clip) {
	var body struct{ Voice string }
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&body); err != nil {
		http.Error(w, "нужен voice", http.StatusBadRequest)
		return
	}
	if _, ok := voiceByID(body.Voice); !ok {
		http.NotFound(w, r)
		return
	}
	_, text, err := s.cands.List(c)
	if err != nil {
		httpError(w, err)
		return
	}
	choice := Choice{Voice: body.Voice, Text: text}
	rec, err := s.lib.Choose(c, s.cands.Path(c, body.Voice), choice)
	if err != nil {
		httpError(w, err)
		return
	}
	log.Printf("%s ← %s «%s»", c.ID(), choice.Voice, choice.Text)
	writeJSON(w, map[string]any{"url": recordingURL(c, rec), "chosen": choice})
}

// pregenerate делает варианты всеми голосами заранее — чтобы в админке не ждать.
func pregenerate(p *Page, cands *Candidates, tts *TTS, all bool) error {
	made := 0
	for _, c := range p.Clips() {
		have, text, err := cands.List(c)
		if err != nil {
			return err
		}
		if text == "" {
			text = p.Spoken(c)
		}
		done := map[string]bool{}
		for _, cand := range have {
			done[cand.Voice] = !all
		}
		for _, v := range voices {
			if done[v.ID] {
				continue
			}
			if _, err := cands.Generate(tts, c, v, text); err != nil {
				return fmt.Errorf("%s %s: %w", c.ID(), v.ID, err)
			}
			made++
		}
		fmt.Printf("  %-12s «%s»\n", c.ID(), text)
	}
	fmt.Printf("Готово: %d вариантов.\n", made)
	return nil
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(w).Encode(v)
}

func httpError(w http.ResponseWriter, err error) {
	log.Print(err)
	http.Error(w, "internal error", http.StatusInternalServerError)
}
