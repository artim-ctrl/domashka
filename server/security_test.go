package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testIndex = `<!doctype html><html><head><script>document.documentElement.dataset.theme = "x";</script>
<script type="module" crossorigin src="./assets/app.js"></script></head><body></body></html>`

// prodHandler — как в проде: без админки, с интерфейсом из web/dist и под префиксом.
func prodHandler(t *testing.T, base string) http.Handler {
	t.Helper()
	page, err := loadPage()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	web := filepath.Join(dir, "web")
	must(t, os.MkdirAll(filepath.Join(web, "assets"), 0o755))
	must(t, os.WriteFile(filepath.Join(web, "index.html"), []byte(testIndex), 0o644))
	must(t, os.WriteFile(filepath.Join(web, "assets", "app.js"), []byte("1"), 0o644))
	words := filepath.Join(dir, "audio", "words")
	must(t, os.MkdirAll(words, 0o755))
	must(t, os.WriteFile(filepath.Join(words, "дом.m4a"), []byte("m4a"), 0o644))

	lib := &Library{dir: filepath.Join(dir, "audio")}
	must(t, lib.Freeze())
	h, err := (&server{page: page, lib: lib}).handler(web, base)
	if err != nil {
		t.Fatal(err)
	}
	return h
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func get(h http.Handler, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", path, nil))
	return rec
}

func TestSecurityHeaders(t *testing.T) {
	h := prodHandler(t, "")
	res := get(h, "/")
	sum := sha256.Sum256([]byte(`document.documentElement.dataset.theme = "x";`))
	csp := res.Header().Get("Content-Security-Policy")
	if !strings.Contains(csp, "script-src 'self' 'sha256-"+base64.StdEncoding.EncodeToString(sum[:])+"'") {
		t.Errorf("CSP не пропускает скрипт темы по хэшу: %s", csp)
	}
	for _, want := range []string{"frame-ancestors 'none'", "object-src 'none'", "style-src 'self'"} {
		if !strings.Contains(csp, want) {
			t.Errorf("в CSP нет %q", want)
		}
	}
	for name, want := range map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"Referrer-Policy":        "no-referrer",
		"Cache-Control":          "no-cache",
	} {
		if got := res.Header().Get(name); got != want {
			t.Errorf("%s: %q, ждали %q", name, got, want)
		}
	}
	if got := get(h, "/api/page").Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Errorf("у API нет nosniff: %q", got)
	}
	if got := get(h, "/assets/app.js").Header().Get("Cache-Control"); !strings.Contains(got, "immutable") {
		t.Errorf("ассеты не кэшируются: %q", got)
	}
}

func TestNoDirectoryListing(t *testing.T) {
	h := prodHandler(t, "")
	if code := get(h, "/assets/").Code; code != http.StatusNotFound {
		t.Errorf("/assets/: %d, ждали 404", code)
	}
}

func TestNoAdminInProd(t *testing.T) {
	h := prodHandler(t, "")
	for _, path := range []string{"/api/admin/voices", "/api/admin/clips/w/%D0%B4%D0%BE%D0%BC"} {
		if code := get(h, path).Code; code != http.StatusNotFound {
			t.Errorf("%s: %d, ждали 404", path, code)
		}
	}
	var resp map[string]any
	must(t, json.NewDecoder(get(h, "/api/page").Body).Decode(&resp))
	for _, key := range []string{"admin", "clips", "choices"} {
		if _, ok := resp[key]; ok {
			t.Errorf("/api/page без админки отдаёт %q", key)
		}
	}
}

func TestBasePath(t *testing.T) {
	h := prodHandler(t, "/domashka")

	res := get(h, "/domashka")
	if res.Code != http.StatusMovedPermanently || res.Header().Get("Location") != "/domashka/" {
		t.Errorf("/domashka: %d → %q, ждали 301 → /domashka/", res.Code, res.Header().Get("Location"))
	}
	if res := get(h, "/domashka/"); res.Code != 200 || !strings.Contains(res.Body.String(), "./assets/app.js") {
		t.Errorf("/domashka/: %d", res.Code)
	}
	if code := get(h, "/domashka/assets/app.js").Code; code != 200 {
		t.Errorf("ассеты под префиксом: %d", code)
	}

	var resp pageResponse
	must(t, json.NewDecoder(get(h, "/domashka/api/page").Body).Decode(&resp))
	url := resp.Recordings["w:дом"]
	if !strings.HasPrefix(url, "api/audio/w/") {
		t.Fatalf("адрес записи должен быть относительным: %q", url)
	}
	if code := get(h, "/domashka/"+url).Code; code != 200 {
		t.Errorf("запись по относительному адресу: %d", code)
	}

	for _, path := range []string{"/", "/api/page"} {
		if code := get(h, path).Code; code != http.StatusNotFound {
			t.Errorf("%s вне префикса: %d, ждали 404", path, code)
		}
	}
	if code := get(h, "/healthz").Code; code != 200 {
		t.Errorf("/healthz: %d", code)
	}
}

func TestFrozenLibrary(t *testing.T) {
	dir := t.TempDir()
	words := filepath.Join(dir, "words")
	must(t, os.MkdirAll(words, 0o755))
	lib := &Library{dir: dir}
	must(t, lib.Freeze())
	must(t, os.WriteFile(filepath.Join(words, "дом.m4a"), nil, 0o644))
	if _, ok, _ := lib.Lookup(Clip{Kind: KindWord, Key: "дом"}); ok {
		t.Error("после Freeze папка читается заново")
	}
}

// Чужой сайт, открытый в том же браузере, не должен дотягиваться до локальной админки.
func TestAdminGuard(t *testing.T) {
	page, err := loadPage()
	must(t, err)
	dir := t.TempDir()
	s := &server{page: page, lib: &Library{dir: dir}, cands: &Candidates{dir: dir}, tts: &TTS{}, admin: true}
	h, err := s.handler("", "")
	must(t, err)

	// Неизвестный голос: если запрос прошёл защиту, ответ — 404.
	const path = "/api/admin/clips/w/%D0%B4%D0%BE%D0%BC/choose"
	for name, tc := range map[string]struct {
		host, contentType string
		headers           map[string]string
		want              int
	}{
		"свой, через Vite":    {"localhost:5180", "application/json", map[string]string{"Sec-Fetch-Site": "same-origin", "Origin": "http://localhost:5180"}, 404},
		"свой, напрямую":      {"127.0.0.1:8080", "application/json", nil, 404},
		"DNS rebinding":       {"evil.example:8080", "application/json", nil, 403},
		"чужой сайт":          {"localhost:8080", "application/json", map[string]string{"Sec-Fetch-Site": "cross-site"}, 403},
		"чужой Origin":        {"localhost:8080", "application/json", map[string]string{"Origin": "https://evil.example"}, 403},
		"форма без preflight": {"localhost:8080", "text/plain", nil, 415},
		"без Content-Type":    {"localhost:8080", "", nil, 415},
	} {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest("POST", path, strings.NewReader(`{"voice":"nobody"}`))
			req.Host = tc.host
			if tc.contentType != "" {
				req.Header.Set("Content-Type", tc.contentType)
			}
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if rec.Code != tc.want {
				t.Errorf("%d, ждали %d (%s)", rec.Code, tc.want, strings.TrimSpace(rec.Body.String()))
			}
		})
	}

	req := httptest.NewRequest("GET", "/api/page", nil)
	req.Host = "evil.example"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Errorf("GET с чужим Host при админке: %d, ждали 403", rec.Code)
	}
}
