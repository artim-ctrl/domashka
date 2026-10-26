package httpserver

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/artim-ctrl/domashka/internal/application"
	"github.com/artim-ctrl/domashka/internal/domain/clip"
	"github.com/artim-ctrl/domashka/internal/domain/content"
	"github.com/artim-ctrl/domashka/internal/domain/voiceover"
	"github.com/artim-ctrl/domashka/internal/infrastructure/audioconv"
	"github.com/artim-ctrl/domashka/internal/infrastructure/candidates"
	"github.com/artim-ctrl/domashka/internal/infrastructure/contentfile"
	"github.com/artim-ctrl/domashka/internal/infrastructure/recordings"
	"github.com/artim-ctrl/domashka/internal/infrastructure/tts"
	"github.com/artim-ctrl/domashka/internal/transport/httpserver/endpoint/audio"
	"github.com/artim-ctrl/domashka/internal/transport/httpserver/endpoint/candidate"
	"github.com/artim-ctrl/domashka/internal/transport/httpserver/endpoint/choose"
	"github.com/artim-ctrl/domashka/internal/transport/httpserver/endpoint/clipstate"
	"github.com/artim-ctrl/domashka/internal/transport/httpserver/endpoint/generate"
	"github.com/artim-ctrl/domashka/internal/transport/httpserver/endpoint/health"
	"github.com/artim-ctrl/domashka/internal/transport/httpserver/endpoint/page"
	"github.com/artim-ctrl/domashka/internal/transport/httpserver/endpoint/voices"
)

const (
	testIndex = `<!doctype html><html><head><script>document.documentElement.dataset.theme = "x";</script>
<script type="module" crossorigin src="./assets/app.js"></script></head><body></body></html>`
	testMaxTextRunes = 200
	testBodyLimit    = 4096
)

var testKindDirs = clip.Dirs{Sounds: "sounds", Words: "words"}

func loadPage(t *testing.T) *content.Page {
	t.Helper()
	p, err := contentfile.Load("../../../content/page.json", "../../../content/speech.json")
	must(t, err)
	return p
}

func recordingOptions(dir string) recordings.Options {
	return recordings.Options{
		Dir:          dir,
		KindDirs:     testKindDirs,
		ChoicesFile:  "choices.json",
		RecordingExt: ".m4a",
		Formats:      []recordings.Format{{Ext: ".m4a", ContentType: "audio/mp4"}, {Ext: ".wav", ContentType: "audio/wav"}},
	}
}

func testOptions(admin bool, base, staticDir string) Options {
	return Options{
		Addr:           "127.0.0.1:0",
		BasePath:       base,
		StaticDir:      staticDir,
		Admin:          admin,
		ReadTimeout:    time.Second,
		WriteTimeout:   time.Second,
		IdleTimeout:    time.Second,
		ReadBufferSize: 64 << 10,
		BodyLimit:      testBodyLimit,
	}
}

func prodServer(t *testing.T, base string) *fiber.App {
	t.Helper()
	dir := t.TempDir()
	web := filepath.Join(dir, "web")
	must(t, os.MkdirAll(filepath.Join(web, "assets"), 0o755))
	must(t, os.WriteFile(filepath.Join(web, "index.html"), []byte(testIndex), 0o644))
	must(t, os.WriteFile(filepath.Join(web, "assets", "app.js"), []byte("1"), 0o644))
	words := filepath.Join(dir, "audio", "words")
	must(t, os.MkdirAll(words, 0o755))
	must(t, os.WriteFile(filepath.Join(words, "дом.m4a"), []byte("m4a"), 0o644))

	store := recordings.New(recordingOptions(filepath.Join(dir, "audio")))
	must(t, store.Freeze())
	pages := application.NewPageService(loadPage(t), store)
	srv, err := New(testOptions(false, base, web), Endpoints{
		Health: health.New(),
		Page:   page.New(pages, nil),
		Audio:  audio.New(pages),
	})
	must(t, err)
	return srv.App()
}

func adminServer(t *testing.T, dir string) *fiber.App {
	t.Helper()
	conv := audioconv.Converter{WAVSampleRate: 24000, AACBitrate: 64000}
	catalog, err := voiceover.NewCatalog([]voiceover.Voice{
		{ID: "macos-milena", Label: "macOS · Milena", License: "macOS", Engine: voiceover.EngineSay, Model: "Milena"},
	})
	must(t, err)
	synth := tts.NewSynthesizer(tts.Options{SileroCommand: []string{"false"}, SayRate: 150, Converter: conv})
	t.Cleanup(synth.Close)

	recOpts := recordingOptions(filepath.Join(dir, "audio"))
	recOpts.Encoder = conv
	store := recordings.New(recOpts)
	cands := candidates.New(candidates.Options{
		Dir:      filepath.Join(dir, "candidates"),
		KindDirs: testKindDirs,
		TextFile: "text.txt",
		Ext:      ".wav",
	})
	pg := loadPage(t)
	pages := application.NewPageService(pg, store)
	svc := application.NewVoiceoverService(pg, catalog, store, cands, synth, application.VoiceoverOptions{MaxTextRunes: testMaxTextRunes})

	srv, err := New(testOptions(true, "", ""), Endpoints{
		Health: health.New(),
		Page:   page.New(pages, svc),
		Audio:  audio.New(pages),
		Admin: &AdminEndpoints{
			Voices:    voices.New(svc),
			ClipState: clipstate.New(svc),
			Generate:  generate.New(svc),
			Choose:    choose.New(svc),
			Candidate: candidate.New(svc),
		},
	})
	must(t, err)
	return srv.App()
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

type result struct {
	Code   int
	Header http.Header
	Body   string
}

func do(t *testing.T, app *fiber.App, req *http.Request) result {
	t.Helper()
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 30 * time.Second})
	must(t, err)
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	must(t, err)
	return result{Code: resp.StatusCode, Header: resp.Header, Body: string(b)}
}

func get(t *testing.T, app *fiber.App, path string) result {
	t.Helper()
	return do(t, app, httptest.NewRequest(http.MethodGet, path, nil))
}

func decode(t *testing.T, r result, v any) {
	t.Helper()
	must(t, json.NewDecoder(strings.NewReader(r.Body)).Decode(v))
}
