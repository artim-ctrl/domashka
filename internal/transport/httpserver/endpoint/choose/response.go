package choose

import (
	"github.com/artim-ctrl/domashka/internal/domain/clip"
	"github.com/artim-ctrl/domashka/internal/domain/voiceover"
	"github.com/artim-ctrl/domashka/internal/transport/httpserver/urls"
)

type Response struct {
	URL    string           `json:"url"`
	Chosen voiceover.Choice `json:"chosen"`
}

func newResponse(c clip.Clip, rec voiceover.Recording, choice voiceover.Choice) Response {
	return Response{URL: urls.Recording(c, rec.Version), Chosen: choice}
}
