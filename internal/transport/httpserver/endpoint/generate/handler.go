package generate

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
	Generate(ctx context.Context, c clip.Clip, voiceID, text string) (voiceover.Candidate, error)
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
	cand, err := h.svc.Generate(c.Context(), cl, req.Voice, req.Text)
	if err != nil {
		if httperr.Status(err) != fiber.StatusInternalServerError {
			return httperr.Handle(c, err)
		}
		slog.Error("generate candidate", "clip", cl.ID(), "voice", req.Voice, "err", err)
		return httperr.Write(c, fiber.StatusBadGateway, err.Error())
	}
	c.Set(fiber.HeaderCacheControl, "no-store")
	return c.JSON(newResponse(cl, cand))
}
