package middleware

import "github.com/gofiber/fiber/v3"

func SecurityHeaders(csp string) fiber.Handler {
	return func(c fiber.Ctx) error {
		c.Set(fiber.HeaderContentSecurityPolicy, csp)
		c.Set(fiber.HeaderXContentTypeOptions, "nosniff")
		c.Set(fiber.HeaderXFrameOptions, "DENY")
		c.Set(fiber.HeaderReferrerPolicy, "no-referrer")
		c.Set("Cross-Origin-Opener-Policy", "same-origin")
		c.Set(fiber.HeaderCrossOriginResourcePolicy, "same-origin")
		c.Set(fiber.HeaderPermissionsPolicy, "camera=(), microphone=(), geolocation=(), payment=(), usb=()")
		return c.Next()
	}
}
