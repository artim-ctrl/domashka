package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const (
	repoBase       = "../../../config/base.yaml"
	repoServe      = "../../../config/serve.yaml"
	repoProduction = "../../../config/production.yaml"
)

func TestRepoConfigsAreValid(t *testing.T) {
	layers := [][]string{{repoBase}, {repoBase, repoServe}, {repoBase, repoProduction}}
	for i := range layers {
		if _, err := Load(layers[i]); err != nil {
			t.Errorf("%v: %v", layers[i], err)
		}
	}
}

func TestLayering(t *testing.T) {
	cfg, err := Load([]string{repoBase, repoProduction})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Server.BasePath != "" || cfg.Server.StaticDir != "/app/web" || cfg.Library.Dir != "/app/audio" {
		t.Errorf("production overrides not applied: %+v", cfg.Server)
	}
	if cfg.Server.Timeouts.Shutdown != 10*time.Second || cfg.Library.ChoicesFile != "choices.json" {
		t.Errorf("base values lost after layering: %+v", cfg)
	}
	if len(cfg.TTS.Voices) == 0 {
		t.Error("voices lost after layering")
	}
}

func TestRejectsUnknownKeys(t *testing.T) {
	path := writeFile(t, "server:\n  adress: 127.0.0.1:8080\n")
	if _, err := Load([]string{repoBase, path}); err == nil || !strings.Contains(err.Error(), "adress") {
		t.Errorf("want unknown key error, got %v", err)
	}
}

func TestValidation(t *testing.T) {
	cases := map[string]struct {
		yaml    string
		wantErr string
	}{
		"bad base path":                {"server:\n  base_path: domashka\n", "server.base_path must start with /"},
		"zero timeout":                 {"server:\n  timeouts:\n    read: 0s\n", "server.timeouts.read must be a positive duration"},
		"ext without dot":              {"library:\n  recording_ext: m4a\n", "library.recording_ext must start with a dot"},
		"same kind dirs":               {"library:\n  kind_dirs:\n    words: sounds\n", "must differ"},
		"unknown engine":               {"tts:\n  voices:\n    - {id: a, label: A, license: MIT, engine: espeak, model: x}\n", "engine must be silero or say"},
		"duplicated voice":             {"tts:\n  voices:\n    - {id: a, label: A, license: MIT, engine: say, model: x}\n    - {id: a, label: A, license: MIT, engine: say, model: x}\n", "is duplicated"},
		"recording ext not in formats": {"library:\n  formats:\n    - {ext: .mp3, content_type: audio/mpeg}\n", `must include ".m4a"`},
		"silero no speaker":            {"tts:\n  voices:\n    - {id: a, label: A, license: MIT, engine: silero, model: x}\n", "speaker is required"},
	}
	for name := range cases {
		tc := cases[name]
		t.Run(name, func(t *testing.T) {
			_, err := Load([]string{repoBase, writeFile(t, tc.yaml)})
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("want %q, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestNoFiles(t *testing.T) {
	if _, err := Load(nil); err == nil {
		t.Error("Load without files succeeded")
	}
}

func writeFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "override.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}
