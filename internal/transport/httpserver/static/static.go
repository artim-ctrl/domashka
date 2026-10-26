package static

import (
	"crypto/sha256"
	"encoding/base64"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v3"
	fiberstatic "github.com/gofiber/fiber/v3/middleware/static"
)

const (
	CacheForever = "public, max-age=31536000, immutable"
	assetsPrefix = "/assets/"
)

func Handler(dir string) fiber.Handler {
	files := fiberstatic.New(dir, fiberstatic.Config{
		ModifyResponse: func(c fiber.Ctx) error {
			if strings.Contains(c.Path(), assetsPrefix) {
				c.Set(fiber.HeaderCacheControl, CacheForever)
			} else {
				c.Set(fiber.HeaderCacheControl, "no-cache")
			}
			return nil
		},
	})
	return func(c fiber.Ctx) error {
		if rel := c.Params("*"); rel != "" && strings.HasSuffix(rel, "/") {
			return fiber.ErrNotFound
		}
		return files(c)
	}
}

var inlineScript = regexp.MustCompile(`(?s)<script>(.*?)</script>`)

func ContentSecurityPolicy(staticDir string) (string, error) {
	scripts := "'self'"
	if staticDir != "" {
		html, err := os.ReadFile(filepath.Join(staticDir, "index.html"))
		if err != nil {
			return "", err
		}
		matches := inlineScript.FindAllSubmatch(html, -1)
		for i := range matches {
			sum := sha256.Sum256(matches[i][1])
			scripts += " 'sha256-" + base64.StdEncoding.EncodeToString(sum[:]) + "'"
		}
	}
	return strings.Join([]string{
		"default-src 'self'",
		"script-src " + scripts,
		"style-src 'self' 'unsafe-inline'",
		"img-src 'self' data:",
		"media-src 'self'",
		"connect-src 'self'",
		"object-src 'none'",
		"base-uri 'none'",
		"form-action 'none'",
		"frame-ancestors 'none'",
	}, "; "), nil
}
