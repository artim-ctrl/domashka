package voices

import "github.com/artim-ctrl/domashka/internal/domain/voiceover"

type Voice struct {
	ID      string `json:"id"`
	Label   string `json:"label"`
	License string `json:"license"`
}

type Response []Voice

func newResponse(voices []voiceover.Voice) Response {
	resp := make(Response, 0, len(voices))
	for i := range voices {
		resp = append(resp, Voice{ID: voices[i].ID, Label: voices[i].Label, License: voices[i].License})
	}
	return resp
}
