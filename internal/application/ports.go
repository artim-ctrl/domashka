package application

import (
	"context"

	"github.com/artim-ctrl/domashka/internal/domain/clip"
	"github.com/artim-ctrl/domashka/internal/domain/voiceover"
)

type RecordingStore interface {
	Scan() (map[clip.Clip]voiceover.Recording, error)
	Lookup(c clip.Clip) (voiceover.Recording, bool, error)
	ExpectedPath(c clip.Clip) string
	ContentType(path string) string
	Choices() (map[string]voiceover.Choice, error)
	Choose(ctx context.Context, c clip.Clip, wav string, choice voiceover.Choice) (voiceover.Recording, error)
}

type CandidateStore interface {
	List(c clip.Clip, voices []voiceover.Voice) ([]voiceover.Candidate, error)
	Text(c clip.Clip) (string, error)
	Get(c clip.Clip, voiceID string) (voiceover.Candidate, bool, error)
	Path(c clip.Clip, voiceID string) string
	SaveText(c clip.Clip, text string) error
}

type Synthesizer interface {
	Synthesize(ctx context.Context, v voiceover.Voice, text, outWAV string) error
}
