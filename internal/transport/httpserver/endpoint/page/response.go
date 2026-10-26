package page

import (
	"github.com/artim-ctrl/domashka/internal/application"
	"github.com/artim-ctrl/domashka/internal/domain/content"
	"github.com/artim-ctrl/domashka/internal/domain/voiceover"
	"github.com/artim-ctrl/domashka/internal/transport/httpserver/urls"
)

type Response struct {
	Page       *content.Page     `json:"page"`
	Recordings map[string]string `json:"recordings"`
	Total      int               `json:"total"`

	Admin   bool                        `json:"admin,omitempty"`
	Clips   []string                    `json:"clips,omitempty"`
	Choices map[string]voiceover.Choice `json:"choices,omitempty"`
}

func newResponse(view application.PageView, admin bool, choices map[string]voiceover.Choice) Response {
	resp := Response{
		Page:       view.Page,
		Recordings: make(map[string]string, len(view.Recordings)),
		Total:      len(view.Clips),
	}
	for c := range view.Recordings {
		resp.Recordings[c.ID()] = urls.Recording(c, view.Recordings[c].Version)
	}
	if admin {
		resp.Admin = true
		resp.Clips = make([]string, 0, len(view.Clips))
		for i := range view.Clips {
			resp.Clips = append(resp.Clips, view.Clips[i].ID())
		}
		resp.Choices = choices
	}
	return resp
}
