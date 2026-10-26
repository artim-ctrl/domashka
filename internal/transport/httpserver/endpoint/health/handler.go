package health

import "github.com/gofiber/fiber/v3"

type Handler struct{}

func New() *Handler { return &Handler{} }

func (h *Handler) Handle(c fiber.Ctx) error {
	return c.JSON(Response{Status: statusOK})
}
