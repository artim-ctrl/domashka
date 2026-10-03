package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func newTestServer(t *testing.T, canRecord bool) (*http.ServeMux, string) {
	t.Helper()
	page, err := loadPage()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	s := &server{page: page, lib: &Library{dir: dir}, canRecord: canRecord}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/clips", s.handleClips)
	mux.HandleFunc("PUT /api/audio/{kind}/{key}", s.handleUpload)
	return mux, dir
}

func put(mux http.Handler, kind, key, contentType, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("PUT", "/api/audio/"+kind+"/"+url.PathEscape(key), strings.NewReader(body))
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

func TestUploadNeedsRecordFlag(t *testing.T) {
	mux, _ := newTestServer(t, false)
	if code := put(mux, "w", "дом", "audio/wav", "RIFF").Code; code != http.StatusForbidden {
		t.Fatalf("без -record: %d, ждали 403", code)
	}
}

func TestUploadOnlyKnownClips(t *testing.T) {
	mux, dir := newTestServer(t, true)
	for _, key := range []string{"кошка", "../../etc/passwd", "дом/../../x"} {
		if code := put(mux, "w", key, "audio/wav", "RIFF").Code; code != http.StatusNotFound {
			t.Errorf("%q: %d, ждали 404", key, code)
		}
	}
	if code := put(mux, "w", "дом", "text/plain", "hi").Code; code != http.StatusUnsupportedMediaType {
		t.Errorf("text/plain: %d, ждали 415", code)
	}
	if entries, _ := os.ReadDir(filepath.Join(dir, "words")); len(entries) != 0 {
		t.Errorf("остались файлы: %v", entries)
	}
}

func TestUploadReplacesOtherFormats(t *testing.T) {
	mux, dir := newTestServer(t, true)
	words := filepath.Join(dir, "words")
	if err := os.MkdirAll(words, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"Щётка.m4a", "щётка.mp3", "шуба.m4a"} {
		if err := os.WriteFile(filepath.Join(words, name), []byte("old"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	rec := put(mux, "w", "щётка", "audio/wav", "RIFF-new")
	if rec.Code != http.StatusOK {
		t.Fatalf("PUT: %d %s", rec.Code, rec.Body)
	}

	var names []string
	entries, _ := os.ReadDir(words)
	for _, e := range entries {
		names = append(names, e.Name())
	}
	if strings.Join(names, ",") != "шуба.m4a,щётка.wav" {
		t.Errorf("в папке %v, ждали шуба.m4a и щётка.wav", names)
	}
	if b, _ := os.ReadFile(filepath.Join(words, "щётка.wav")); string(b) != "RIFF-new" {
		t.Errorf("содержимое %q", b)
	}

	req := httptest.NewRequest("GET", "/api/clips", nil)
	got := httptest.NewRecorder()
	mux.ServeHTTP(got, req)
	var resp struct {
		Clips []struct{ Kind, Key, URL string }
	}
	if err := json.NewDecoder(got.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	for _, c := range resp.Clips {
		if c.Key == "щётка" && !strings.HasPrefix(c.URL, "/api/audio/w/") {
			t.Errorf("у щётки нет адреса записи: %+v", c)
		}
		if c.Key == "дом" && c.URL != "" {
			t.Errorf("у дома не должно быть записи: %+v", c)
		}
	}
}
