package bootstrap

import (
	"path/filepath"
	"testing"

	"go.uber.org/fx"

	"github.com/artim-ctrl/domashka/internal/infrastructure/config"
)

func repoConfig(t *testing.T) config.Config {
	t.Helper()
	root := "../.."
	cfg, err := config.Load([]string{filepath.Join(root, "config/base.yaml")})
	if err != nil {
		t.Fatal(err)
	}
	cfg.Content.Page = filepath.Join(root, cfg.Content.Page)
	cfg.Content.Speech = filepath.Join(root, cfg.Content.Speech)
	cfg.Library.Dir = t.TempDir()
	cfg.Candidates.Dir = t.TempDir()
	return cfg
}

func TestGraphs(t *testing.T) {
	cfg := repoConfig(t)
	cases := map[string][]fx.Option{
		"serve":       {coreModule(cfg, Runtime{}), httpModule(Runtime{})},
		"serve admin": {coreModule(cfg, Runtime{Admin: true}), httpModule(Runtime{Admin: true}), voiceoverModule()},
		"missing":     {coreModule(cfg, Runtime{})},
		"voices":      {coreModule(cfg, Runtime{Admin: true}), voiceoverModule()},
	}
	for name := range cases {
		opts := append([]fx.Option{fx.NopLogger}, cases[name]...)
		if err := fx.ValidateApp(opts...); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
}

func TestAdminServeNeedsVoiceoverModule(t *testing.T) {
	cfg := repoConfig(t)
	err := fx.ValidateApp(fx.NopLogger, coreModule(cfg, Runtime{Admin: true}), httpModule(Runtime{Admin: true}))
	if err == nil {
		t.Error("admin server graph resolved without the voiceover module")
	}
}
