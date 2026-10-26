package httpserver

import (
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/artim-ctrl/domashka/internal/transport/httpserver/endpoint/page"
)

func TestSecurityHeaders(t *testing.T) {
	app := prodServer(t, "")
	res := get(t, app, "/")
	sum := sha256.Sum256([]byte(`document.documentElement.dataset.theme = "x";`))
	csp := res.Header.Get("Content-Security-Policy")
	if !strings.Contains(csp, "script-src 'self' 'sha256-"+base64.StdEncoding.EncodeToString(sum[:])+"'") {
		t.Errorf("CSP does not allow the theme script by hash: %s", csp)
	}
	directives := []string{"frame-ancestors 'none'", "object-src 'none'", "style-src 'self' 'unsafe-inline'"}
	for i := range directives {
		if !strings.Contains(csp, directives[i]) {
			t.Errorf("CSP lacks %q", directives[i])
		}
	}
	policies := strings.Split(csp, "; ")
	for i := range policies {
		if strings.HasPrefix(policies[i], "script-src") && strings.Contains(policies[i], "unsafe") {
			t.Errorf("script-src must not allow unsafe sources: %s", policies[i])
		}
	}
	headers := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"Referrer-Policy":        "no-referrer",
		"Cache-Control":          "no-cache",
	}
	for name := range headers {
		if got := res.Header.Get(name); got != headers[name] {
			t.Errorf("%s: %q, want %q", name, got, headers[name])
		}
	}
	if got := get(t, app, "/api/page").Header.Get("X-Content-Type-Options"); got != "nosniff" {
		t.Errorf("API response lacks nosniff: %q", got)
	}
	if got := get(t, app, "/assets/app.js").Header.Get("Cache-Control"); !strings.Contains(got, "immutable") {
		t.Errorf("assets are not cached: %q", got)
	}
}

func TestNoDirectoryListing(t *testing.T) {
	app := prodServer(t, "")
	paths := []string{"/assets/", "/assets"}
	for i := range paths {
		if code := get(t, app, paths[i]).Code; code != http.StatusNotFound {
			t.Errorf("%s: %d, want 404", paths[i], code)
		}
	}
}

func TestNoAdminInProd(t *testing.T) {
	app := prodServer(t, "")
	paths := []string{"/api/admin/voices", "/api/admin/clips/w/%D0%B4%D0%BE%D0%BC"}
	for i := range paths {
		if code := get(t, app, paths[i]).Code; code != http.StatusNotFound {
			t.Errorf("%s: %d, want 404", paths[i], code)
		}
	}
	var resp map[string]any
	decode(t, get(t, app, "/api/page"), &resp)
	keys := []string{"admin", "clips", "choices"}
	for i := range keys {
		if _, ok := resp[keys[i]]; ok {
			t.Errorf("/api/page without admin mode returns %q", keys[i])
		}
	}
}

func TestAudioContentType(t *testing.T) {
	app := prodServer(t, "")
	var resp page.Response
	decode(t, get(t, app, "/api/page"), &resp)
	res := get(t, app, "/"+resp.Recordings["w:дом"])
	if res.Code != http.StatusOK || res.Header.Get("Content-Type") != "audio/mp4" {
		t.Errorf("recording: %d %q, want 200 audio/mp4", res.Code, res.Header.Get("Content-Type"))
	}
	if code := get(t, app, "/api/audio/x/%D0%B4%D0%BE%D0%BC").Code; code != http.StatusNotFound {
		t.Errorf("unknown kind: %d, want 404", code)
	}
}

func TestBasePath(t *testing.T) {
	app := prodServer(t, "/domashka")

	res := get(t, app, "/domashka")
	if res.Code != http.StatusMovedPermanently || res.Header.Get("Location") != "/domashka/" {
		t.Errorf("/domashka: %d → %q, want 301 → /domashka/", res.Code, res.Header.Get("Location"))
	}
	if res := get(t, app, "/domashka/"); res.Code != http.StatusOK || !strings.Contains(res.Body, "./assets/app.js") {
		t.Errorf("/domashka/: %d", res.Code)
	}
	if code := get(t, app, "/domashka/assets/app.js").Code; code != http.StatusOK {
		t.Errorf("assets under the prefix: %d", code)
	}

	var resp page.Response
	decode(t, get(t, app, "/domashka/api/page"), &resp)
	recURL := resp.Recordings["w:дом"]
	if !strings.HasPrefix(recURL, "api/audio/w/") {
		t.Fatalf("recording URL must be relative: %q", recURL)
	}
	if code := get(t, app, "/domashka/"+recURL).Code; code != http.StatusOK {
		t.Errorf("recording by relative URL: %d", code)
	}

	paths := []string{"/", "/api/page"}
	for i := range paths {
		if code := get(t, app, paths[i]).Code; code != http.StatusNotFound {
			t.Errorf("%s outside the prefix: %d, want 404", paths[i], code)
		}
	}
	if code := get(t, app, "/healthz").Code; code != http.StatusOK {
		t.Errorf("/healthz: %d", code)
	}
}

func TestAdminGuard(t *testing.T) {
	app := adminServer(t, t.TempDir())

	const path = "/api/admin/clips/w/%D0%B4%D0%BE%D0%BC/choose"
	cases := map[string]struct {
		host, contentType string
		headers           map[string]string
		want              int
	}{
		"own page via Vite":      {"localhost:5180", "application/json", map[string]string{"Sec-Fetch-Site": "same-origin", "Origin": "http://localhost:5180"}, http.StatusNotFound},
		"own page directly":      {"127.0.0.1:8080", "application/json", nil, http.StatusNotFound},
		"DNS rebinding":          {"evil.example:8080", "application/json", nil, http.StatusForbidden},
		"cross-site":             {"localhost:8080", "application/json", map[string]string{"Sec-Fetch-Site": "cross-site"}, http.StatusForbidden},
		"foreign Origin":         {"localhost:8080", "application/json", map[string]string{"Origin": "https://evil.example"}, http.StatusForbidden},
		"form without preflight": {"localhost:8080", "text/plain", nil, http.StatusUnsupportedMediaType},
		"no Content-Type":        {"localhost:8080", "", nil, http.StatusUnsupportedMediaType},
	}
	for name := range cases {
		tc := cases[name]
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{"voice":"nobody"}`))
			req.Host = tc.host
			if tc.contentType != "" {
				req.Header.Set("Content-Type", tc.contentType)
			}
			for k := range tc.headers {
				req.Header.Set(k, tc.headers[k])
			}
			if res := do(t, app, req); res.Code != tc.want {
				t.Errorf("%d, want %d (%s)", res.Code, tc.want, strings.TrimSpace(res.Body))
			}
		})
	}

	req := httptest.NewRequest(http.MethodGet, "/api/page", nil)
	req.Host = "evil.example"
	if res := do(t, app, req); res.Code != http.StatusForbidden {
		t.Errorf("GET with a foreign Host in admin mode: %d, want 403", res.Code)
	}
}

func TestAdminNeedsEndpoints(t *testing.T) {
	if _, err := New(testOptions(true, "", ""), Endpoints{}); err == nil {
		t.Error("admin server built without admin endpoints")
	}
}
