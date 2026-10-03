package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:8080", "адрес HTTP-сервера")
	audioDir := flag.String("audio", "../audio", "папка с записями (внутри sounds/ и words/)")
	staticDir := flag.String("static", "", "собранный интерфейс (web/dist); пусто — не раздавать")
	record := flag.Bool("record", false, "разрешить запись из браузера (страница /record)")
	missing := flag.Bool("missing", false, "показать, каких записей не хватает, и выйти")
	flag.Parse()

	page, err := loadPage()
	if err != nil {
		log.Fatal(err)
	}
	lib := &Library{dir: *audioDir}

	if *missing {
		if err := printMissing(page, lib); err != nil {
			log.Fatal(err)
		}
		return
	}

	s := &server{page: page, lib: lib, canRecord: *record}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/page", s.handlePage)
	mux.HandleFunc("GET /api/clips", s.handleClips)
	mux.HandleFunc("GET /api/audio/{kind}/{key}", s.handleAudio)
	mux.HandleFunc("PUT /api/audio/{kind}/{key}", s.handleUpload)
	if *staticDir != "" {
		mux.Handle("GET /", spa(*staticDir))
	}

	mode := "только чтение"
	if *record {
		mode = "запись разрешена"
	}
	log.Printf("domashka слушает http://%s (записи: %s, %s)", *addr, *audioDir, mode)
	log.Fatal(http.ListenAndServe(*addr, mux))
}

type server struct {
	page      *Page
	lib       *Library
	canRecord bool
}

func recordingURL(c Clip, rec Recording) string {
	return fmt.Sprintf("/api/audio/%s/%s?v=%d", c.Kind, url.PathEscape(c.Key), rec.Version)
}

type pageResponse struct {
	Page *Page `json:"page"`
	// Recordings: «s:ш» / «w:дом» → адрес записи. Чего нет в словаре — ещё не записано.
	Recordings map[string]string `json:"recordings"`
	Total      int               `json:"total"`
	CanRecord  bool              `json:"canRecord"`
}

func (s *server) handlePage(w http.ResponseWriter, r *http.Request) {
	have, err := s.lib.Scan()
	if err != nil {
		httpError(w, err)
		return
	}
	clips := s.page.Clips()
	resp := pageResponse{Page: s.page, Recordings: map[string]string{}, Total: len(clips), CanRecord: s.canRecord}
	for _, c := range clips {
		if rec, ok := have[Clip{Kind: c.Kind, Key: normalizeKey(c.Key)}]; ok {
			resp.Recordings[c.ID()] = recordingURL(c, rec)
		}
	}
	writeJSON(w, resp)
}

type clipItem struct {
	Kind   Kind     `json:"kind"`
	Key    string   `json:"key"`
	URL    string   `json:"url,omitempty"` // пусто — ещё не записано
	IPA    *Text    `json:"ipa,omitempty"`
	UsedIn []string `json:"usedIn"`
}

type clipsResponse struct {
	CanRecord bool            `json:"canRecord"`
	Sections  map[string]Text `json:"sections"` // id → заголовок, для подписей «где звучит»
	Clips     []clipItem      `json:"clips"`
}

func (s *server) handleClips(w http.ResponseWriter, r *http.Request) {
	have, err := s.lib.Scan()
	if err != nil {
		httpError(w, err)
		return
	}
	resp := clipsResponse{CanRecord: s.canRecord, Sections: map[string]Text{}, Clips: []clipItem{}}
	for _, sec := range s.page.Sections {
		if sec.Title != nil {
			resp.Sections[sec.ID] = *sec.Title
		}
	}
	for _, u := range s.page.ClipUses() {
		item := clipItem{Kind: u.Kind, Key: u.Key, IPA: u.IPA, UsedIn: u.Sections}
		if rec, ok := have[Clip{Kind: u.Kind, Key: normalizeKey(u.Key)}]; ok {
			item.URL = recordingURL(u.Clip, rec)
		}
		resp.Clips = append(resp.Clips, item)
	}
	writeJSON(w, resp)
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

// Что принимаем на запись. Страница /record сама присылает WAV, остальное — на всякий случай.
var uploadExts = map[string]string{
	"audio/wav": ".wav", "audio/x-wav": ".wav", "audio/wave": ".wav",
	"audio/mp4": ".m4a", "audio/mpeg": ".mp3", "audio/ogg": ".ogg", "audio/webm": ".webm",
}

const maxUpload = 20 << 20

func (s *server) handleUpload(w http.ResponseWriter, r *http.Request) {
	if !s.canRecord {
		http.Error(w, "запись выключена: запусти сервер с -record", http.StatusForbidden)
		return
	}
	// Писать можно только записи, которые есть на странице, — путь к файлу строится из её ключа.
	clip, ok := s.page.FindClip(Kind(r.PathValue("kind")), r.PathValue("key"))
	if !ok {
		http.NotFound(w, r)
		return
	}
	mt, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	ext, ok := uploadExts[mt]
	if !ok {
		http.Error(w, fmt.Sprintf("неподходящий формат %q", mt), http.StatusUnsupportedMediaType)
		return
	}
	rec, err := s.lib.Save(clip, ext, http.MaxBytesReader(w, r.Body, maxUpload))
	var tooBig *http.MaxBytesError
	if errors.As(err, &tooBig) {
		http.Error(w, "файл слишком большой", http.StatusRequestEntityTooLarge)
		return
	}
	if err != nil {
		httpError(w, err)
		return
	}
	log.Printf("записано %s → %s", clip.ID(), rec.Path)
	writeJSON(w, map[string]string{"url": recordingURL(clip, rec)})
}

// spa раздаёт собранный интерфейс; адреса страниц (/record) получают index.html.
func spa(dir string) http.Handler {
	files := http.FileServer(http.Dir(dir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(r.URL.Path))); errors.Is(err, os.ErrNotExist) {
			http.ServeFile(w, r, filepath.Join(dir, "index.html"))
			return
		}
		files.ServeHTTP(w, r)
	})
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
