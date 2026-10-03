package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckSpeech(t *testing.T) {
	p, err := loadPage()
	if err != nil {
		t.Fatal(err)
	}
	for name, tc := range map[string]struct {
		edit    func(map[string]string)
		wantErr string
	}{
		"пропущено слово":    {func(m map[string]string) { delete(m, "w:дом") }, "нет текста для w:дом"},
		"опечатка в слове":   {func(m map[string]string) { m["w:дом"] = "д+оом" }, "это не то же слово"},
		"согласная одна":     {func(m map[string]string) { m["s:ш"] = "ш" }, "прочитает название буквы"},
		"лишняя запись":      {func(m map[string]string) { m["w:кошка"] = "к+ошка" }, "w:кошка нет на странице"},
		"ударение не мешает": {func(m map[string]string) { m["w:дом"] = "д+ом" }, ""},
	} {
		t.Run(name, func(t *testing.T) {
			q := *p
			q.speech = map[string]string{}
			for k, v := range p.speech {
				q.speech[k] = v
			}
			tc.edit(q.speech)
			err := q.checkSpeech()
			if tc.wantErr == "" && err != nil || tc.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tc.wantErr)) {
				t.Errorf("ждали %q, получили %v", tc.wantErr, err)
			}
		})
	}
}

// Полный цикл админки на голосе macOS: сгенерировать вариант, выбрать, увидеть на странице.
func TestAdminGenerateAndChoose(t *testing.T) {
	for _, tool := range []string{"say", "afconvert"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skip("нужен macOS:", tool)
		}
	}
	page, err := loadPage()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	s := &server{page: page, lib: &Library{dir: filepath.Join(dir, "audio")}, cands: &Candidates{dir: filepath.Join(dir, "cand")}, tts: &TTS{}, admin: true}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/page", s.handlePage)
	mux.HandleFunc("GET /api/admin/clips/{kind}/{key}", s.withClip(s.handleClip))
	mux.HandleFunc("POST /api/admin/clips/{kind}/{key}/candidates/{voice}", s.withClip(s.handleGenerate))
	mux.HandleFunc("POST /api/admin/clips/{kind}/{key}/choose", s.withClip(s.handleChoose))
	call := func(method, path, body string, out any) int {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(method, path, strings.NewReader(body)))
		if out != nil && rec.Code == http.StatusOK {
			if err := json.NewDecoder(rec.Body).Decode(out); err != nil {
				t.Fatal(err)
			}
		}
		return rec.Code
	}
	clip := "/api/admin/clips/w/" + url.PathEscape("молоко")

	var state clipResponse
	if call("GET", clip, "", &state) != 200 || state.Text != "молок+о" || len(state.Candidates) != 0 {
		t.Fatalf("до генерации: %+v", state)
	}
	if code := call("POST", "/api/admin/clips/w/"+url.PathEscape("кошка")+"/candidates/macos-milena", `{"text":"к+ошка"}`, nil); code != 404 {
		t.Errorf("слово не со страницы: %d", code)
	}
	if code := call("POST", clip+"/candidates/nobody", `{"text":"молок+о"}`, nil); code != 404 {
		t.Errorf("неизвестный голос: %d", code)
	}

	var cand Candidate
	if code := call("POST", clip+"/candidates/macos-milena", `{"text":"молок+о"}`, &cand); code != 200 || cand.Voice != "macos-milena" {
		t.Fatalf("генерация: %d %+v", code, cand)
	}
	var chosen struct {
		URL    string
		Chosen Choice
	}
	if code := call("POST", clip+"/choose", `{"voice":"macos-milena"}`, &chosen); code != 200 || chosen.Chosen.Text != "молок+о" {
		t.Fatalf("выбор: %d %+v", code, chosen)
	}
	if _, err := os.Stat(filepath.Join(dir, "audio", "words", "молоко.m4a")); err != nil {
		t.Errorf("запись не появилась: %v", err)
	}

	var pageResp pageResponse
	call("GET", "/api/page", "", &pageResp)
	if pageResp.Recordings["w:молоко"] != chosen.URL || pageResp.Choices["w:молоко"].Voice != "macos-milena" || len(pageResp.Clips) != pageResp.Total {
		t.Errorf("страница не видит выбор: %v %v", pageResp.Recordings["w:молоко"], pageResp.Choices)
	}
}
