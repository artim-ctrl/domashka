package middleware

import (
	"mime"
	"net"
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v3"

	"github.com/artim-ctrl/domashka/internal/transport/httpserver/httperr"
)

func AdminGuard() fiber.Handler {
	return func(c fiber.Ctx) error {
		if !isLoopback(c.Host()) {
			return httperr.Write(c, fiber.StatusForbidden, "the admin mode is available on localhost only")
		}
		if isCrossSite(c) {
			return httperr.Write(c, fiber.StatusForbidden, "cross-site request")
		}
		if c.Method() != fiber.MethodGet && c.Method() != fiber.MethodHead {
			if mt, _, _ := mime.ParseMediaType(c.Get(fiber.HeaderContentType)); mt != fiber.MIMEApplicationJSON {
				return httperr.Write(c, fiber.StatusUnsupportedMediaType, "Content-Type must be application/json")
			}
		}
		return c.Next()
	}
}

func isCrossSite(c fiber.Ctx) bool {
	if site := c.Get("Sec-Fetch-Site"); site != "" && site != "same-origin" && site != "none" {
		return true
	}
	if origin := c.Get(fiber.HeaderOrigin); origin != "" {
		u, err := url.Parse(origin)
		return err != nil || !isLoopback(u.Host)
	}
	return false
}

func isLoopback(hostport string) bool {
	host := hostport
	if h, _, err := net.SplitHostPort(hostport); err == nil {
		host = h
	}
	host = strings.Trim(host, "[]")
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
