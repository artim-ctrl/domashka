package page

import (
	"github.com/gofiber/fiber/v3"

	"github.com/artim-ctrl/domashka/internal/application"
	"github.com/artim-ctrl/domashka/internal/domain/voiceover"
	"github.com/artim-ctrl/domashka/internal/transport/httpserver/httperr"
)

type Viewer interface {
	View() (application.PageView, error)
}

type ChoiceSource interface {
	Choices() (map[string]voiceover.Choice, error)
}

type Handler struct {
	pages   Viewer
	choices ChoiceSource
}

func New(pages Viewer, choices ChoiceSource) *Handler {
	return &Handler{pages: pages, choices: choices}
}

func (h *Handler) Handle(c fiber.Ctx) error {
	view, err := h.pages.View()
	if err != nil {
		return httperr.Handle(c, err)
	}
	var choices map[string]voiceover.Choice
	if h.choices != nil {
		if choices, err = h.choices.Choices(); err != nil {
			return httperr.Handle(c, err)
		}
	}
	c.Set(fiber.HeaderCacheControl, "no-store")
	return c.JSON(newResponse(view, h.choices != nil, choices))
}
