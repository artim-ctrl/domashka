package bootstrap

import (
	"context"
	"fmt"
	"io"

	"go.uber.org/fx"

	"github.com/artim-ctrl/domashka/internal/application"
	"github.com/artim-ctrl/domashka/internal/infrastructure/config"
)

func Run(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	f, err := parseFlags(args, stderr)
	if err != nil {
		return err
	}
	cfg, err := config.Load(f.configs)
	if err != nil {
		return err
	}

	switch f.mode {
	case modeHealthcheck:
		return checkHealth(ctx, cfg.Server.Addr, cfg.Server.Timeouts.Healthcheck)
	case modeMissing:
		return runMissing(ctx, cfg, stdout)
	case modeVoices, modeVoicesAll:
		return runPregenerate(ctx, cfg, stdout, f.mode == modeVoicesAll)
	default:
		return runServer(ctx, cfg, Runtime{Admin: f.admin})
	}
}

func runServer(ctx context.Context, cfg config.Config, rt Runtime) error {
	opts := []fx.Option{fxLogger(), coreModule(cfg, rt), httpModule(rt)}
	if rt.Admin {
		opts = append(opts, voiceoverModule())
	}
	app := fx.New(append(opts, fx.StartTimeout(cfg.Server.Timeouts.Startup), fx.StopTimeout(cfg.Server.Timeouts.Shutdown))...)
	if err := app.Err(); err != nil {
		return err
	}
	if err := startApp(ctx, app, cfg); err != nil {
		return err
	}
	<-ctx.Done()
	return stopApp(ctx, app, cfg)
}

func runMissing(ctx context.Context, cfg config.Config, out io.Writer) error {
	var pages *application.PageService
	app := fx.New(fxLogger(), coreModule(cfg, Runtime{}), fx.Populate(&pages))
	if err := app.Err(); err != nil {
		return err
	}
	recorded, missing, err := pages.Missing()
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "Recorded %d of %d.\n", recorded, recorded+len(missing))
	for i := range missing {
		fmt.Fprintln(out, "  "+missing[i])
	}
	return nil
}

func runPregenerate(ctx context.Context, cfg config.Config, out io.Writer, all bool) error {
	var svc *application.VoiceoverService
	rt := Runtime{Admin: true}
	app := fx.New(fxLogger(), coreModule(cfg, rt), voiceoverModule(), fx.Populate(&svc))
	if err := app.Err(); err != nil {
		return err
	}
	if err := startApp(ctx, app, cfg); err != nil {
		return err
	}
	made, runErr := svc.Pregenerate(ctx, out, all)
	stopErr := stopApp(ctx, app, cfg)
	if runErr != nil {
		return runErr
	}
	fmt.Fprintf(out, "Done: %d candidates.\n", made)
	return stopErr
}

func startApp(ctx context.Context, app *fx.App, cfg config.Config) error {
	startCtx, cancel := context.WithTimeout(ctx, cfg.Server.Timeouts.Startup)
	defer cancel()
	return app.Start(startCtx)
}

func stopApp(ctx context.Context, app *fx.App, cfg config.Config) error {
	stopCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), cfg.Server.Timeouts.Shutdown)
	defer cancel()
	return app.Stop(stopCtx)
}
