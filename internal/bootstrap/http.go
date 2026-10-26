package bootstrap

import (
	"go.uber.org/fx"

	"github.com/artim-ctrl/domashka/internal/application"
	"github.com/artim-ctrl/domashka/internal/infrastructure/config"
	"github.com/artim-ctrl/domashka/internal/transport/httpserver"
	"github.com/artim-ctrl/domashka/internal/transport/httpserver/endpoint/audio"
	"github.com/artim-ctrl/domashka/internal/transport/httpserver/endpoint/candidate"
	"github.com/artim-ctrl/domashka/internal/transport/httpserver/endpoint/choose"
	"github.com/artim-ctrl/domashka/internal/transport/httpserver/endpoint/clipstate"
	"github.com/artim-ctrl/domashka/internal/transport/httpserver/endpoint/generate"
	"github.com/artim-ctrl/domashka/internal/transport/httpserver/endpoint/health"
	"github.com/artim-ctrl/domashka/internal/transport/httpserver/endpoint/page"
	"github.com/artim-ctrl/domashka/internal/transport/httpserver/endpoint/voices"
)

type pageHandlerParams struct {
	fx.In

	Pages     *application.PageService
	Voiceover *application.VoiceoverService `optional:"true"`
}

func newPageHandler(p pageHandlerParams) *page.Handler {
	if p.Voiceover == nil {
		return page.New(p.Pages, nil)
	}
	return page.New(p.Pages, p.Voiceover)
}

func newAdminEndpoints(svc *application.VoiceoverService) *httpserver.AdminEndpoints {
	return &httpserver.AdminEndpoints{
		Voices:    voices.New(svc),
		ClipState: clipstate.New(svc),
		Generate:  generate.New(svc),
		Choose:    choose.New(svc),
		Candidate: candidate.New(svc),
	}
}

type serverParams struct {
	fx.In

	Lifecycle fx.Lifecycle
	Config    config.Config
	Runtime   Runtime
	Health    *health.Handler
	Page      *page.Handler
	Audio     *audio.Handler
	Admin     *httpserver.AdminEndpoints `optional:"true"`
}

func newServer(p serverParams) (*httpserver.Server, error) {
	srv := p.Config.Server
	writeTimeout := srv.Timeouts.Write
	if p.Runtime.Admin {
		writeTimeout = srv.Timeouts.AdminWrite
	}
	server, err := httpserver.New(httpserver.Options{
		Addr:           srv.Addr,
		BasePath:       srv.BasePath,
		StaticDir:      srv.StaticDir,
		Admin:          p.Runtime.Admin,
		ReadTimeout:    srv.Timeouts.Read,
		WriteTimeout:   writeTimeout,
		IdleTimeout:    srv.Timeouts.Idle,
		ReadBufferSize: srv.ReadBufferSize,
		BodyLimit:      srv.BodyLimit,
	}, httpserver.Endpoints{
		Health: p.Health,
		Page:   p.Page,
		Audio:  p.Audio,
		Admin:  p.Admin,
	})
	if err != nil {
		return nil, err
	}
	p.Lifecycle.Append(fx.StartStopHook(server.Start, server.Stop))
	return server, nil
}

func httpModule(rt Runtime) fx.Option {
	opts := []fx.Option{
		fx.Provide(
			health.New,
			newPageHandler,
			fx.Annotate(audio.New, fx.From(new(*application.PageService))),
			newServer,
		),
		fx.Invoke(func(*httpserver.Server) {}),
	}
	if rt.Admin {
		opts = append(opts, fx.Provide(newAdminEndpoints), fx.Invoke(func(*httpserver.AdminEndpoints) {}))
	}
	return fx.Module("http", opts...)
}
