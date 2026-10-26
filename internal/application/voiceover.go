package application

import (
	"context"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/artim-ctrl/domashka/internal/domain/clip"
	"github.com/artim-ctrl/domashka/internal/domain/content"
	"github.com/artim-ctrl/domashka/internal/domain/voiceover"
)

type ClipState struct {
	Text       string
	Candidates []voiceover.Candidate
	Chosen     *voiceover.Choice
}

type CandidateFile struct {
	Path        string
	ContentType string
}

type VoiceoverOptions struct {
	MaxTextRunes int
}

type VoiceoverService struct {
	page       *content.Page
	voices     *voiceover.Catalog
	recordings RecordingStore
	candidates CandidateStore
	synth      Synthesizer
	opts       VoiceoverOptions
}

func NewVoiceoverService(
	page *content.Page,
	voices *voiceover.Catalog,
	recordings RecordingStore,
	candidates CandidateStore,
	synth Synthesizer,
	opts VoiceoverOptions,
) *VoiceoverService {
	return &VoiceoverService{
		page:       page,
		voices:     voices,
		recordings: recordings,
		candidates: candidates,
		synth:      synth,
		opts:       opts,
	}
}

func (s *VoiceoverService) Voices() []voiceover.Voice { return s.voices.All() }

func (s *VoiceoverService) Choices() (map[string]voiceover.Choice, error) {
	return s.recordings.Choices()
}

func (s *VoiceoverService) FindClip(kind, key string) (clip.Clip, error) {
	k, ok := clip.ParseKind(kind)
	if !ok {
		return clip.Clip{}, fmt.Errorf("clip kind %q: %w", kind, ErrNotFound)
	}
	c, ok := s.page.FindClip(k, key)
	if !ok {
		return clip.Clip{}, fmt.Errorf("clip %s:%s is not on the page: %w", kind, key, ErrNotFound)
	}
	return c, nil
}

func (s *VoiceoverService) ClipState(c clip.Clip) (ClipState, error) {
	cands, err := s.candidates.List(c, s.voices.All())
	if err != nil {
		return ClipState{}, err
	}
	text, err := s.candidates.Text(c)
	if err != nil {
		return ClipState{}, err
	}
	choices, err := s.recordings.Choices()
	if err != nil {
		return ClipState{}, err
	}
	if text == "" {
		text = s.page.Spoken(c)
	}
	state := ClipState{Text: text, Candidates: cands}
	if ch, ok := choices[c.ID()]; ok {
		state.Chosen = &ch
	}
	return state, nil
}

func (s *VoiceoverService) Generate(ctx context.Context, c clip.Clip, voiceID, text string) (voiceover.Candidate, error) {
	v, err := s.voice(voiceID)
	if err != nil {
		return voiceover.Candidate{}, err
	}
	text = strings.TrimSpace(text)
	switch {
	case text == "":
		return voiceover.Candidate{}, fmt.Errorf("text is required: %w", ErrInvalid)
	case utf8.RuneCountInString(text) > s.opts.MaxTextRunes:
		return voiceover.Candidate{}, fmt.Errorf("text is longer than %d characters: %w", s.opts.MaxTextRunes, ErrInvalid)
	}
	return s.generate(ctx, c, v, text)
}

func (s *VoiceoverService) CandidateFile(c clip.Clip, voiceID string) (CandidateFile, error) {
	v, err := s.voice(voiceID)
	if err != nil {
		return CandidateFile{}, err
	}
	path := s.candidates.Path(c, v.ID)
	return CandidateFile{Path: path, ContentType: s.recordings.ContentType(path)}, nil
}

func (s *VoiceoverService) Choose(ctx context.Context, c clip.Clip, voiceID string) (voiceover.Recording, voiceover.Choice, error) {
	v, err := s.voice(voiceID)
	if err != nil {
		return voiceover.Recording{}, voiceover.Choice{}, err
	}
	if _, ok, err := s.candidates.Get(c, v.ID); err != nil {
		return voiceover.Recording{}, voiceover.Choice{}, err
	} else if !ok {
		return voiceover.Recording{}, voiceover.Choice{}, fmt.Errorf("no candidate generated with voice %q yet: %w", v.ID, ErrConflict)
	}
	text, err := s.candidates.Text(c)
	if err != nil {
		return voiceover.Recording{}, voiceover.Choice{}, err
	}
	choice := voiceover.Choice{Voice: v.ID, Text: text}
	rec, err := s.recordings.Choose(ctx, c, s.candidates.Path(c, v.ID), choice)
	if err != nil {
		return voiceover.Recording{}, voiceover.Choice{}, err
	}
	return rec, choice, nil
}

func (s *VoiceoverService) Pregenerate(ctx context.Context, out io.Writer, all bool) (int, error) {
	voices := s.voices.All()
	clips := s.page.Clips()
	made := 0
	for i := range clips {
		c := clips[i]
		have, err := s.candidates.List(c, voices)
		if err != nil {
			return made, err
		}
		text, err := s.candidates.Text(c)
		if err != nil {
			return made, err
		}
		if text == "" {
			text = s.page.Spoken(c)
		}
		done := make(map[string]bool, len(have))
		for j := range have {
			done[have[j].Voice] = !all
		}
		for j := range voices {
			if done[voices[j].ID] {
				continue
			}
			if err := ctx.Err(); err != nil {
				return made, err
			}
			if _, err := s.generate(ctx, c, voices[j], text); err != nil {
				return made, fmt.Errorf("%s %s: %w", c.ID(), voices[j].ID, err)
			}
			made++
		}
		fmt.Fprintf(out, "  %-12s %q\n", c.ID(), text)
	}
	return made, nil
}

func (s *VoiceoverService) voice(id string) (voiceover.Voice, error) {
	v, ok := s.voices.ByID(id)
	if !ok {
		return voiceover.Voice{}, fmt.Errorf("unknown voice %q: %w", id, ErrNotFound)
	}
	return v, nil
}

func (s *VoiceoverService) generate(ctx context.Context, c clip.Clip, v voiceover.Voice, text string) (voiceover.Candidate, error) {
	if err := s.candidates.SaveText(c, text); err != nil {
		return voiceover.Candidate{}, err
	}
	if err := s.synth.Synthesize(ctx, v, text, s.candidates.Path(c, v.ID)); err != nil {
		return voiceover.Candidate{}, err
	}
	cand, ok, err := s.candidates.Get(c, v.ID)
	if err != nil {
		return voiceover.Candidate{}, err
	}
	if !ok {
		return voiceover.Candidate{}, fmt.Errorf("synthesizer did not write a candidate for %s %s", c.ID(), v.ID)
	}
	return cand, nil
}
