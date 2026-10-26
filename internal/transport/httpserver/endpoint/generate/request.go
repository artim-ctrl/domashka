package generate

import (
	"github.com/gofiber/fiber/v3"

	"github.com/artim-ctrl/domashka/internal/transport/httpserver/params"
)

type Request struct {
	params.ClipPath
	Voice string `json:"-"`
	Text  string `json:"text"`
}

func parseRequest(c fiber.Ctx) (Request, error) {
	req := Request{ClipPath: params.ParseClipPath(c), Voice: c.Params("voice")}
	if err := params.DecodeJSON(c, &req); err != nil {
		return Request{}, err
	}
	return req, nil
}
