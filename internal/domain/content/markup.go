package content

import (
	"regexp"

	"github.com/artim-ctrl/domashka/internal/domain/clip"
)

var clipToken = regexp.MustCompile(`\{\{([sw]):([^}|]+)(?:\|[^}]*)?\}\}`)

func clipsIn(s string) []clip.Clip {
	matches := clipToken.FindAllStringSubmatch(s, -1)
	out := make([]clip.Clip, 0, len(matches))
	for i := range matches {
		out = append(out, clip.Clip{Kind: clip.Kind(matches[i][1]), Key: matches[i][2]})
	}
	return out
}

func tokens(s string) []string {
	clips := clipsIn(s)
	out := make([]string, 0, len(clips))
	for i := range clips {
		out = append(out, clips[i].ID())
	}
	return out
}
