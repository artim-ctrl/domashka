package httpserver

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"strings"

	"github.com/gofiber/fiber/v3"

	"github.com/artim-ctrl/domashka/internal/transport/httpserver/httperr"
	"github.com/artim-ctrl/domashka/internal/transport/httpserver/middleware"
	"github.com/artim-ctrl/domashka/internal/transport/httpserver/static"
)

type Server struct {
	app  *fiber.App
	opts Options
}

func New(opts Options, endpoints Endpoints) (*Server, error) {
	opts.BasePath = strings.TrimRight(opts.BasePath, "/")
	csp, err := static.ContentSecurityPolicy(opts.StaticDir)
	if err != nil {
		return nil, err
	}

	app := fiber.New(fiber.Config{
		StrictRouting:  true,
		CaseSensitive:  true,
		UnescapePath:   true,
		ReadTimeout:    opts.ReadTimeout,
		WriteTimeout:   opts.WriteTimeout,
		IdleTimeout:    opts.IdleTimeout,
		ReadBufferSize: opts.ReadBufferSize,
		BodyLimit:      opts.BodyLimit,
		ErrorHandler:   httperr.Handle,
	})
	app.Use(middleware.SecurityHeaders(csp))
	if opts.Admin {
		app.Use(middleware.AdminGuard())
	}
	if err := registerRoutes(app, opts, endpoints); err != nil {
		return nil, err
	}
	return &Server{app: app, opts: opts}, nil
}

func (s *Server) App() *fiber.App { return s.app }

func (s *Server) Start(ctx context.Context) error {
	ln, err := (&net.ListenConfig{}).Listen(ctx, "tcp", s.opts.Addr)
	if err != nil {
		return err
	}
	slog.Info("listening", "url", "http://"+s.opts.Addr+s.opts.BasePath+"/", "admin", s.opts.Admin)
	go func() {
		if err := s.app.Listener(ln, fiber.ListenConfig{DisableStartupMessage: true}); err != nil && !errors.Is(err, net.ErrClosed) {
			slog.Error("http server stopped", "err", err)
		}
	}()
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	slog.Info("shutting down")
	return s.app.ShutdownWithContext(ctx)
}
