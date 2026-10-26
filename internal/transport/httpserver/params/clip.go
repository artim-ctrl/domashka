package params

import "github.com/gofiber/fiber/v3"

type ClipPath struct {
	Kind string
	Key  string
}

func ParseClipPath(c fiber.Ctx) ClipPath {
	return ClipPath{Kind: c.Params("kind"), Key: c.Params("key")}
}
