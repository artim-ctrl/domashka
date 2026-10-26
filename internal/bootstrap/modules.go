package bootstrap

import (
	"log/slog"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"

	"github.com/artim-ctrl/domashka/internal/application"
	"github.com/artim-ctrl/domashka/internal/infrastructure/config"
)

func coreModule(cfg config.Config, rt Runtime) fx.Option {
	return fx.Module("core",
		fx.Supply(cfg, rt),
		fx.Provide(
			newConverter,
			loadPage,
			newRecordingStore,
			application.NewPageService,
		),
	)
}

func voiceoverModule() fx.Option {
	return fx.Module("voiceover",
		fx.Provide(
			newCatalog,
			newSynthesizer,
			newCandidateStore,
			newVoiceoverService,
		),
	)
}

func fxLogger() fx.Option {
	return fx.WithLogger(func() fxevent.Logger {
		logger := &fxevent.SlogLogger{Logger: slog.Default()}
		logger.UseLogLevel(slog.LevelDebug)
		return logger
	})
}
