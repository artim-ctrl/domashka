package params

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/gofiber/fiber/v3"

	"github.com/artim-ctrl/domashka/internal/application"
)

func DecodeJSON(c fiber.Ctx, v any) error {
	dec := json.NewDecoder(bytes.NewReader(c.Body()))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return fmt.Errorf("invalid JSON body: %v: %w", err, application.ErrInvalid)
	}
	return nil
}
