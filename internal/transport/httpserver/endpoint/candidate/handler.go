package candidate

import (
	"github.com/gofiber/fiber/v3"

	"github.com/artim-ctrl/domashka/internal/application"
	"github.com/artim-ctrl/domashka/internal/domain/clip"
	"github.com/artim-ctrl/domashka/internal/transport/httpserver/httperr"
	"github.com/artim-ctrl/domashka/internal/transport/httpserver/static"
)

type Service interface {
	FindClip(kind, key string) (clip.Clip, error)
	CandidateFile(c clip.Clip, voiceID string) (application.CandidateFile, error)
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
	file, err := h.svc.CandidateFile(cl, req.Voice)
	if err != nil {
		return httperr.Handle(c, err)
	}
	if err := c.SendFile(file.Path); err != nil {
		return err
	}
	c.Set(fiber.HeaderContentType, file.ContentType)
	c.Set(fiber.HeaderCacheControl, static.CacheForever)
	return nil
}
