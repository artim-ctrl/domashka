package httpserver

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"

	"github.com/artim-ctrl/domashka/internal/transport/httpserver/endpoint/choose"
	"github.com/artim-ctrl/domashka/internal/transport/httpserver/endpoint/clipstate"
	"github.com/artim-ctrl/domashka/internal/transport/httpserver/endpoint/generate"
	"github.com/artim-ctrl/domashka/internal/transport/httpserver/endpoint/page"
)

func call(t *testing.T, app *fiber.App, method, path, body string, out any) int {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Host = "localhost:8080"
	req.Header.Set("Content-Type", "application/json")
	res := do(t, app, req)
	if out != nil && res.Code == http.StatusOK {
		decode(t, res, out)
	}
	return res.Code
}

func TestAdminRejectsBadInput(t *testing.T) {
	app := adminServer(t, t.TempDir())
	clipPath := "/api/admin/clips/w/" + url.PathEscape("молоко")

	cases := map[string]struct {
		method, path, body string
		want               int
	}{
		"word not on the page":   {http.MethodPost, "/api/admin/clips/w/" + url.PathEscape("кошка") + "/candidates/macos-milena", `{"text":"к+ошка"}`, http.StatusNotFound},
		"unknown kind":           {http.MethodGet, "/api/admin/clips/x/" + url.PathEscape("молоко"), "", http.StatusNotFound},
		"unknown voice":          {http.MethodPost, clipPath + "/candidates/nobody", `{"text":"молок+о"}`, http.StatusNotFound},
		"empty text":             {http.MethodPost, clipPath + "/candidates/macos-milena", `{"text":"  "}`, http.StatusBadRequest},
		"text too long":          {http.MethodPost, clipPath + "/candidates/macos-milena", `{"text":"` + strings.Repeat("а", testMaxTextRunes+1) + `"}`, http.StatusBadRequest},
		"unknown field":          {http.MethodPost, clipPath + "/candidates/macos-milena", `{"txt":"молок+о"}`, http.StatusBadRequest},
		"choose before generate": {http.MethodPost, clipPath + "/choose", `{"voice":"macos-milena"}`, http.StatusConflict},
	}
	for name := range cases {
		tc := cases[name]
		t.Run(name, func(t *testing.T) {
			if code := call(t, app, tc.method, tc.path, tc.body, nil); code != tc.want {
				t.Errorf("%d, want %d", code, tc.want)
			}
		})
	}
}

func TestAdminGenerateAndChoose(t *testing.T) {
	tools := []string{"say", "afconvert"}
	for i := range tools {
		if _, err := exec.LookPath(tools[i]); err != nil {
			t.Skip("needs macOS:", tools[i])
		}
	}
	dir := t.TempDir()
	app := adminServer(t, dir)
	clipPath := "/api/admin/clips/w/" + url.PathEscape("молоко")

	var state clipstate.Response
	if call(t, app, http.MethodGet, clipPath, "", &state) != http.StatusOK || state.Text != "молок+о" || len(state.Candidates) != 0 {
		t.Fatalf("before generating: %+v", state)
	}

	var cand generate.Response
	if code := call(t, app, http.MethodPost, clipPath+"/candidates/macos-milena", `{"text":"молок+о"}`, &cand); code != http.StatusOK || cand.Voice != "macos-milena" {
		t.Fatalf("generate: %d %+v", code, cand)
	}
	if code := get(t, app, "/"+cand.URL).Code; code != http.StatusForbidden {
		t.Errorf("candidate without a localhost Host: %d, want 403", code)
	}
	var chosen choose.Response
	if code := call(t, app, http.MethodPost, clipPath+"/choose", `{"voice":"macos-milena"}`, &chosen); code != http.StatusOK || chosen.Chosen.Text != "молок+о" {
		t.Fatalf("choose: %d %+v", code, chosen)
	}
	if _, err := os.Stat(filepath.Join(dir, "audio", "words", "молоко.m4a")); err != nil {
		t.Errorf("recording was not written: %v", err)
	}

	var pg page.Response
	call(t, app, http.MethodGet, "/api/page", "", &pg)
	if pg.Recordings["w:молоко"] != chosen.URL || pg.Choices["w:молоко"].Voice != "macos-milena" || len(pg.Clips) != pg.Total {
		t.Errorf("page does not reflect the choice: %v %v", pg.Recordings["w:молоко"], pg.Choices)
	}
}
