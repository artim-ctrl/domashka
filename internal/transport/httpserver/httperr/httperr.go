package httperr

import (
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v3"

	"github.com/artim-ctrl/domashka/internal/application"
)

type Response struct {
	Error string `json:"error"`
}

func Write(c fiber.Ctx, status int, msg string) error {
	c.Set(fiber.HeaderCacheControl, "no-store")
	return c.Status(status).JSON(Response{Error: msg})
}

func Status(err error) int {
	var fe *fiber.Error
	switch {
	case errors.Is(err, application.ErrNotFound):
		return fiber.StatusNotFound
	case errors.Is(err, application.ErrInvalid):
		return fiber.StatusBadRequest
	case errors.Is(err, application.ErrConflict):
		return fiber.StatusConflict
	case errors.As(err, &fe):
		return fe.Code
	default:
		return fiber.StatusInternalServerError
	}
}

func Handle(c fiber.Ctx, err error) error {
	status := Status(err)
	if status == fiber.StatusInternalServerError {
		slog.Error("request failed", "method", c.Method(), "path", c.Path(), "err", err)
		return Write(c, status, "internal error")
	}
	return Write(c, status, err.Error())
}
