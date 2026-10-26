package clipstate

import (
	"github.com/gofiber/fiber/v3"

	"github.com/artim-ctrl/domashka/internal/application"
	"github.com/artim-ctrl/domashka/internal/domain/clip"
	"github.com/artim-ctrl/domashka/internal/transport/httpserver/httperr"
)

type Service interface {
	FindClip(kind, key string) (clip.Clip, error)
	ClipState(c clip.Clip) (application.ClipState, error)
}

type Handler struct {
	svc Service
}

func New(svc Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Handle(c fiber.Ctx) error {
	req := parseRequest(c)
	cl, err := h.svc.FindClip(req.Kind, req.Key)
	if err != nil {
		return httperr.Handle(c, err)
	}
	state, err := h.svc.ClipState(cl)
	if err != nil {
		return httperr.Handle(c, err)
	}
	c.Set(fiber.HeaderCacheControl, "no-store")
	return c.JSON(newResponse(cl, state))
}
