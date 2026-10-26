package choose

import (
	"context"
	"log/slog"

	"github.com/gofiber/fiber/v3"

	"github.com/artim-ctrl/domashka/internal/domain/clip"
	"github.com/artim-ctrl/domashka/internal/domain/voiceover"
	"github.com/artim-ctrl/domashka/internal/transport/httpserver/httperr"
)

type Service interface {
	FindClip(kind, key string) (clip.Clip, error)
	Choose(ctx context.Context, c clip.Clip, voiceID string) (voiceover.Recording, voiceover.Choice, error)
}

type Handler struct {
	svc Service
}

func New(svc Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Handle(c fiber.Ctx) error {
	req, err := parseRequest(c)
	if err != nil {
		return httperr.Handle(c, err)
	}
	cl, err := h.svc.FindClip(req.Kind, req.Key)
	if err != nil {
		return httperr.Handle(c, err)
	}
	rec, choice, err := h.svc.Choose(c.Context(), cl, req.Voice)
	if err != nil {
		return httperr.Handle(c, err)
	}
	slog.Info("recording chosen", "clip", cl.ID(), "voice", choice.Voice, "text", choice.Text)
	c.Set(fiber.HeaderCacheControl, "no-store")
	return c.JSON(newResponse(cl, rec, choice))
}
