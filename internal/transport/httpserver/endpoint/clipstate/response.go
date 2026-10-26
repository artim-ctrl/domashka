package clipstate

import (
	"github.com/artim-ctrl/domashka/internal/application"
	"github.com/artim-ctrl/domashka/internal/domain/clip"
	"github.com/artim-ctrl/domashka/internal/domain/voiceover"
	"github.com/artim-ctrl/domashka/internal/transport/httpserver/urls"
)

type Candidate struct {
	Voice string `json:"voice"`
	URL   string `json:"url"`
}

type Response struct {
	Text       string            `json:"text"`
	Candidates []Candidate       `json:"candidates"`
	Chosen     *voiceover.Choice `json:"chosen,omitempty"`
}

func newResponse(c clip.Clip, state application.ClipState) Response {
	resp := Response{Text: state.Text, Candidates: make([]Candidate, 0, len(state.Candidates)), Chosen: state.Chosen}
	for i := range state.Candidates {
		cand := state.Candidates[i]
		resp.Candidates = append(resp.Candidates, Candidate{Voice: cand.Voice, URL: urls.Candidate(c, cand.Voice, cand.Version)})
	}
	return resp
}
