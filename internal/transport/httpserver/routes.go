package httpserver

import (
	"errors"

	"github.com/gofiber/fiber/v3"

	"github.com/artim-ctrl/domashka/internal/transport/httpserver/endpoint/audio"
	"github.com/artim-ctrl/domashka/internal/transport/httpserver/endpoint/candidate"
	"github.com/artim-ctrl/domashka/internal/transport/httpserver/endpoint/choose"
	"github.com/artim-ctrl/domashka/internal/transport/httpserver/endpoint/clipstate"
	"github.com/artim-ctrl/domashka/internal/transport/httpserver/endpoint/generate"
	"github.com/artim-ctrl/domashka/internal/transport/httpserver/endpoint/health"
	"github.com/artim-ctrl/domashka/internal/transport/httpserver/endpoint/page"
	"github.com/artim-ctrl/domashka/internal/transport/httpserver/endpoint/voices"
	"github.com/artim-ctrl/domashka/internal/transport/httpserver/static"
)

type Endpoints struct {
	Health *health.Handler
	Page   *page.Handler
	Audio  *audio.Handler
	Admin  *AdminEndpoints
}

type AdminEndpoints struct {
	Voices    *voices.Handler
	ClipState *clipstate.Handler
	Generate  *generate.Handler
	Choose    *choose.Handler
	Candidate *candidate.Handler
}

func registerRoutes(app *fiber.App, opts Options, e Endpoints) error {
	if opts.Admin && e.Admin == nil {
		return errors.New("admin mode needs admin endpoints")
	}
	app.Get("/healthz", e.Health.Handle)

	base := opts.BasePath
	if base != "" {
		app.Get(base, func(c fiber.Ctx) error {
			return c.Redirect().Status(fiber.StatusMovedPermanently).To(base + "/")
		})
	}
	root := app.Group(base)
	root.Get("/api/page", e.Page.Handle)
	root.Get("/api/audio/:kind/:key", e.Audio.Handle)

	if opts.Admin {
		admin := root.Group("/api/admin")
		admin.Get("/voices", e.Admin.Voices.Handle)
		admin.Get("/clips/:kind/:key", e.Admin.ClipState.Handle)
		admin.Post("/clips/:kind/:key/candidates/:voice", e.Admin.Generate.Handle)
		admin.Post("/clips/:kind/:key/choose", e.Admin.Choose.Handle)
		admin.Get("/candidates/:kind/:key/:voice", e.Admin.Candidate.Handle)
	}

	if opts.StaticDir != "" {
		root.Get("/*", static.Handler(opts.StaticDir))
	}
	return nil
}
