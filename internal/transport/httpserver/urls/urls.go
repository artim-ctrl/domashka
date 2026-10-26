package urls

import (
	"fmt"
	"net/url"

	"github.com/artim-ctrl/domashka/internal/domain/clip"
)

func Recording(c clip.Clip, version int64) string {
	return fmt.Sprintf("api/audio/%s/%s?v=%d", c.Kind, url.PathEscape(c.Key), version)
}

func Candidate(c clip.Clip, voiceID string, version int64) string {
	return fmt.Sprintf("api/admin/candidates/%s/%s/%s?v=%d", c.Kind, url.PathEscape(c.Key), url.PathEscape(voiceID), version)
}
