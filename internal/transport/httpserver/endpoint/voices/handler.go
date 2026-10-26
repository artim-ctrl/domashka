package voices

import (
	"github.com/gofiber/fiber/v3"

	"github.com/artim-ctrl/domashka/internal/domain/voiceover"
)

type Lister interface {
	Voices() []voiceover.Voice
}

type Handler struct {
	voices Lister
}

func New(voices Lister) *Handler { return &Handler{voices: voices} }

func (h *Handler) Handle(c fiber.Ctx) error {
	c.Set(fiber.HeaderCacheControl, "no-store")
	return c.JSON(newResponse(h.voices.Voices()))
}
