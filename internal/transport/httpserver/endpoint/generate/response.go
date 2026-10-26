package generate

import (
	"github.com/artim-ctrl/domashka/internal/domain/clip"
	"github.com/artim-ctrl/domashka/internal/domain/voiceover"
	"github.com/artim-ctrl/domashka/internal/transport/httpserver/urls"
)

type Response struct {
	Voice string `json:"voice"`
	URL   string `json:"url"`
}

func newResponse(c clip.Clip, cand voiceover.Candidate) Response {
	return Response{Voice: cand.Voice, URL: urls.Candidate(c, cand.Voice, cand.Version)}
}
