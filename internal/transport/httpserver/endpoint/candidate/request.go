package candidate

import (
	"github.com/gofiber/fiber/v3"

	"github.com/artim-ctrl/domashka/internal/transport/httpserver/params"
)

type Request struct {
	params.ClipPath
	Voice string
}

func parseRequest(c fiber.Ctx) Request {
	return Request{ClipPath: params.ParseClipPath(c), Voice: c.Params("voice")}
}
