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

	s := &server{page: page, lib: lib}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/page", s.handlePage)
	mux.HandleFunc("GET /api/audio/{kind}/{key}", s.handleAudio)
	if *staticDir != "" {
		mux.Handle("GET /", http.FileServer(http.Dir(*staticDir)))
	}

	log.Printf("domashka слушает http://%s (записи: %s)", *addr, *audioDir)
	log.Fatal(http.ListenAndServe(*addr, mux))
}

type server struct {
	page *Page
	lib  *Library
}

type pageResponse struct {
	Page *Page `json:"page"`
	// Recordings: «s:ш» / «w:дом» → адрес записи. Чего нет в словаре — ещё не записано.
	Recordings map[string]string `json:"recordings"`
	Total      int               `json:"total"`
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
			resp.Recordings[c.ID()] = fmt.Sprintf("/api/audio/%s/%s?v=%d", c.Kind, url.PathEscape(c.Key), rec.Version)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(w).Encode(resp)
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

func httpError(w http.ResponseWriter, err error) {
	log.Print(err)
	http.Error(w, "internal error", http.StatusInternalServerError)
}
