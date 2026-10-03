package main

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"mime"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// handler собирает всё приложение. base — префикс адреса, под которым оно живёт
// (например, /domashka на example.com/domashka); прокси пересылает путь как есть.
// Интерфейс и API ходят по относительным адресам, так что сборка от префикса не зависит.
func (s *server) handler(staticDir, base string) (http.Handler, error) {
	app := http.NewServeMux()
	app.HandleFunc("GET /api/page", s.handlePage)
	app.HandleFunc("GET /api/audio/{kind}/{key}", s.handleAudio)
	if s.admin {
		app.HandleFunc("GET /api/admin/voices", s.handleVoices)
		app.HandleFunc("GET /api/admin/clips/{kind}/{key}", s.withClip(s.handleClip))
		app.HandleFunc("POST /api/admin/clips/{kind}/{key}/candidates/{voice}", s.withClip(s.handleGenerate))
		app.HandleFunc("POST /api/admin/clips/{kind}/{key}/choose", s.withClip(s.handleChoose))
		app.HandleFunc("GET /api/admin/candidates/{kind}/{key}/{voice}", s.withClip(s.handleCandidate))
	}
	if staticDir != "" {
		app.Handle("GET /", static(staticDir))
	}

	root := http.NewServeMux()
	root.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) { fmt.Fprintln(w, "ok") })
	if base = strings.TrimRight(base, "/"); base == "" {
		root.Handle("/", app)
	} else {
		root.Handle(base+"/", http.StripPrefix(base, app))
		// Без завершающего «/» относительные адреса интерфейса указали бы мимо приложения.
		root.Handle("GET "+base, http.RedirectHandler(base+"/", http.StatusMovedPermanently))
	}

	csp, err := contentSecurityPolicy(staticDir)
	if err != nil {
		return nil, err
	}
	var h http.Handler = root
	if s.admin {
		h = adminGuard(h)
	}
	return securityHeaders(h, csp), nil
}

// static раздаёт собранный интерфейс. У файлов из assets/ в имени хэш содержимого —
// их можно кэшировать навсегда; index.html — всегда проверять, иначе не увидишь новую версию.
// Списков файлов в папках не показываем.
func static(dir string) http.Handler {
	files := http.FileServer(http.Dir(dir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" && strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/assets/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		} else {
			w.Header().Set("Cache-Control", "no-cache")
		}
		files.ServeHTTP(w, r)
	})
}

var inlineScript = regexp.MustCompile(`(?s)<script>(.*?)</script>`)

// contentSecurityPolicy разрешает только своё: скрипты, стили, звук и запросы — с того же
// адреса. Встроенный скрипт темы в index.html пропускается по хэшу — его считаем при старте
// из собранной страницы, чтобы правка скрипта не ломала CSP.
func contentSecurityPolicy(staticDir string) (string, error) {
	scripts := "'self'"
	if staticDir != "" {
		html, err := os.ReadFile(filepath.Join(staticDir, "index.html"))
		if err != nil {
			return "", err
		}
		for _, m := range inlineScript.FindAllSubmatch(html, -1) {
			sum := sha256.Sum256(m[1])
			scripts += " 'sha256-" + base64.StdEncoding.EncodeToString(sum[:]) + "'"
		}
	}
	return strings.Join([]string{
		"default-src 'self'",
		"script-src " + scripts,
		"style-src 'self'",
		"img-src 'self' data:",
		"media-src 'self'",
		"connect-src 'self'",
		"object-src 'none'",
		"base-uri 'none'",
		"form-action 'none'",
		"frame-ancestors 'none'",
	}, "; "), nil
}

func securityHeaders(h http.Handler, csp string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hd := w.Header()
		hd.Set("Content-Security-Policy", csp)
		hd.Set("X-Content-Type-Options", "nosniff")
		hd.Set("X-Frame-Options", "DENY")
		hd.Set("Referrer-Policy", "no-referrer")
		hd.Set("Cross-Origin-Opener-Policy", "same-origin")
		hd.Set("Cross-Origin-Resource-Policy", "same-origin")
		hd.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=(), payment=(), usb=()")
		h.ServeHTTP(w, r)
	})
}

// adminGuard закрывает локальную админку от чужих сайтов, открытых в том же браузере:
//   - Host — только localhost: иначе сайт, чей домен «перепривязан» на 127.0.0.1 (DNS rebinding),
//     ходил бы в админку как к себе домой;
//   - запросы с другого сайта (Sec-Fetch-Site, Origin) — отказ;
//   - изменения — только с Content-Type: application/json: такой запрос с чужой страницы
//     браузер не отправит без предварительного согласования, а его сервер не даёт.
func adminGuard(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isLoopback(r.Host) {
			http.Error(w, "админка доступна только с localhost", http.StatusForbidden)
			return
		}
		if site := r.Header.Get("Sec-Fetch-Site"); site != "" && site != "same-origin" && site != "none" {
			http.Error(w, "запрос с чужого сайта", http.StatusForbidden)
			return
		}
		if origin := r.Header.Get("Origin"); origin != "" {
			if u, err := url.Parse(origin); err != nil || !isLoopback(u.Host) {
				http.Error(w, "запрос с чужого сайта", http.StatusForbidden)
				return
			}
		}
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			if mt, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type")); mt != "application/json" {
				http.Error(w, "нужен Content-Type: application/json", http.StatusUnsupportedMediaType)
				return
			}
		}
		h.ServeHTTP(w, r)
	})
}

func isLoopback(hostport string) bool {
	host := hostport
	if h, _, err := net.SplitHostPort(hostport); err == nil {
		host = h
	}
	host = strings.Trim(host, "[]")
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
