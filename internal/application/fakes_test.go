package application

import (
	"context"
	"os"
	"testing"

	"github.com/artim-ctrl/domashka/internal/domain/clip"
	"github.com/artim-ctrl/domashka/internal/domain/content"
	"github.com/artim-ctrl/domashka/internal/domain/voiceover"
)

type fakeRecordings struct {
	have    map[clip.Clip]voiceover.Recording
	choices map[string]voiceover.Choice
	chosen  []string
}

func (f *fakeRecordings) Scan() (map[clip.Clip]voiceover.Recording, error) { return f.have, nil }

func (f *fakeRecordings) Lookup(c clip.Clip) (voiceover.Recording, bool, error) {
	rec, ok := f.have[c.Normalized()]
	return rec, ok, nil
}

func (f *fakeRecordings) ExpectedPath(c clip.Clip) string { return c.ID() }

func (f *fakeRecordings) ContentType(string) string { return "audio/mp4" }

func (f *fakeRecordings) Choices() (map[string]voiceover.Choice, error) { return f.choices, nil }

func (f *fakeRecordings) Choose(_ context.Context, c clip.Clip, wav string, choice voiceover.Choice) (voiceover.Recording, error) {
	f.chosen = append(f.chosen, wav)
	f.choices[c.ID()] = choice
	return voiceover.Recording{Path: c.ID(), Version: 1}, nil
}

type fakeCandidates struct {
	texts map[clip.Clip]string
	files map[string]bool
}

func (f *fakeCandidates) key(c clip.Clip, voiceID string) string { return c.ID() + "/" + voiceID }

func (f *fakeCandidates) List(c clip.Clip, voices []voiceover.Voice) ([]voiceover.Candidate, error) {
	var out []voiceover.Candidate
	for i := range voices {
		if f.files[f.key(c, voices[i].ID)] {
			out = append(out, voiceover.Candidate{Voice: voices[i].ID, Version: 1})
		}
	}
	return out, nil
}

func (f *fakeCandidates) Text(c clip.Clip) (string, error) { return f.texts[c], nil }

func (f *fakeCandidates) Get(c clip.Clip, voiceID string) (voiceover.Candidate, bool, error) {
	if !f.files[f.key(c, voiceID)] {
		return voiceover.Candidate{}, false, nil
	}
	return voiceover.Candidate{Voice: voiceID, Version: 1}, true, nil
}

func (f *fakeCandidates) Path(c clip.Clip, voiceID string) string { return f.key(c, voiceID) }

func (f *fakeCandidates) SaveText(c clip.Clip, text string) error {
	f.texts[c] = text
	return nil
}

type fakeSynth struct {
	cands *fakeCandidates
}

func (f fakeSynth) Synthesize(_ context.Context, _ voiceover.Voice, _, outWAV string) error {
	f.cands.files[outWAV] = true
	return nil
}

func loadPage(t *testing.T) *content.Page {
	t.Helper()
	pageData, err := os.ReadFile("../../content/page.json")
	if err != nil {
		t.Fatal(err)
	}
	speechData, err := os.ReadFile("../../content/speech.json")
	if err != nil {
		t.Fatal(err)
	}
	p, err := content.Parse(pageData, speechData)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

type fixture struct {
	page  *content.Page
	recs  *fakeRecordings
	cands *fakeCandidates
	svc   *VoiceoverService
}

func newFixture(t *testing.T) fixture {
	t.Helper()
	page := loadPage(t)
	catalog, err := voiceover.NewCatalog([]voiceover.Voice{
		{ID: "a", Engine: voiceover.EngineSay},
		{ID: "b", Engine: voiceover.EngineSay},
	})
	if err != nil {
		t.Fatal(err)
	}
	recs := &fakeRecordings{have: map[clip.Clip]voiceover.Recording{}, choices: map[string]voiceover.Choice{}}
	cands := &fakeCandidates{texts: map[clip.Clip]string{}, files: map[string]bool{}}
	svc := NewVoiceoverService(page, catalog, recs, cands, fakeSynth{cands: cands}, VoiceoverOptions{MaxTextRunes: 10})
	return fixture{page: page, recs: recs, cands: cands, svc: svc}
}
