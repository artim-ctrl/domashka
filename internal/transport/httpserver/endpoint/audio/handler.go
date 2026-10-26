package audio

import (
	"github.com/gofiber/fiber/v3"

	"github.com/artim-ctrl/domashka/internal/application"
	"github.com/artim-ctrl/domashka/internal/transport/httpserver/httperr"
	"github.com/artim-ctrl/domashka/internal/transport/httpserver/static"
)

type Finder interface {
	Audio(kind, key string) (application.AudioFile, error)
}

type Handler struct {
	audio Finder
}

func New(audio Finder) *Handler { return &Handler{audio: audio} }

func (h *Handler) Handle(c fiber.Ctx) error {
	req := parseRequest(c)
	file, err := h.audio.Audio(req.Kind, req.Key)
	if err != nil {
		return httperr.Handle(c, err)
	}
	if err := c.SendFile(file.Recording.Path); err != nil {
		return err
	}
	c.Set(fiber.HeaderContentType, file.ContentType)
	c.Set(fiber.HeaderCacheControl, static.CacheForever)
	return nil
}
