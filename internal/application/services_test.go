package application

import (
	"errors"
	"io"
	"testing"

	"github.com/artim-ctrl/domashka/internal/domain/clip"
	"github.com/artim-ctrl/domashka/internal/domain/voiceover"
)

var dom = clip.Clip{Kind: clip.Word, Key: "дом"}

func TestFindClip(t *testing.T) {
	f := newFixture(t)
	if c, err := f.svc.FindClip("w", "Дом"); err != nil || c != dom {
		t.Errorf("FindClip: %v %v", c, err)
	}
	cases := [][2]string{{"x", "дом"}, {"w", "кошка"}}
	for i := range cases {
		if _, err := f.svc.FindClip(cases[i][0], cases[i][1]); !errors.Is(err, ErrNotFound) {
			t.Errorf("%v: want ErrNotFound, got %v", cases[i], err)
		}
	}
}

func TestClipStateFallsBackToSpeech(t *testing.T) {
	f := newFixture(t)
	state, err := f.svc.ClipState(dom)
	if err != nil {
		t.Fatal(err)
	}
	if state.Text != f.page.Spoken(dom) || len(state.Candidates) != 0 || state.Chosen != nil {
		t.Errorf("state: %+v", state)
	}
}

func TestGenerateValidates(t *testing.T) {
	f := newFixture(t)
	cases := map[string]struct {
		voice, text string
		want        error
	}{
		"unknown voice": {"nobody", "д+ом", ErrNotFound},
		"empty text":    {"a", "   ", ErrInvalid},
		"too long":      {"a", "ддддддддддд", ErrInvalid},
	}
	for name := range cases {
		tc := cases[name]
		if _, err := f.svc.Generate(t.Context(), dom, tc.voice, tc.text); !errors.Is(err, tc.want) {
			t.Errorf("%s: want %v, got %v", name, tc.want, err)
		}
	}
}

func TestGenerateAndChoose(t *testing.T) {
	f := newFixture(t)
	if _, _, err := f.svc.Choose(t.Context(), dom, "a"); !errors.Is(err, ErrConflict) {
		t.Fatalf("choose before generate: want ErrConflict, got %v", err)
	}
	cand, err := f.svc.Generate(t.Context(), dom, "a", " д+ом ")
	if err != nil || cand.Voice != "a" {
		t.Fatalf("generate: %+v %v", cand, err)
	}
	_, choice, err := f.svc.Choose(t.Context(), dom, "a")
	if err != nil || choice != (voiceover.Choice{Voice: "a", Text: "д+ом"}) {
		t.Fatalf("choose: %+v %v", choice, err)
	}
	state, err := f.svc.ClipState(dom)
	if err != nil || state.Chosen == nil || state.Chosen.Voice != "a" || len(state.Candidates) != 1 {
		t.Errorf("state after choose: %+v %v", state, err)
	}
}

func TestPregenerateSkipsExisting(t *testing.T) {
	f := newFixture(t)
	if _, err := f.svc.Generate(t.Context(), dom, "a", "д+ом"); err != nil {
		t.Fatal(err)
	}
	clips := len(f.page.Clips())
	made, err := f.svc.Pregenerate(t.Context(), io.Discard, false)
	if err != nil || made != clips*2-1 {
		t.Errorf("pregenerate: made %d, want %d (%v)", made, clips*2-1, err)
	}
	made, err = f.svc.Pregenerate(t.Context(), io.Discard, true)
	if err != nil || made != clips*2 {
		t.Errorf("pregenerate all: made %d, want %d (%v)", made, clips*2, err)
	}
}

func TestPageService(t *testing.T) {
	f := newFixture(t)
	f.recs.have[dom] = voiceover.Recording{Path: "дом.m4a", Version: 7}
	pages := NewPageService(f.page, f.recs)

	view, err := pages.View()
	if err != nil || view.Recordings[dom].Version != 7 || len(view.Clips) != len(f.page.Clips()) {
		t.Errorf("view: %v %v", view.Recordings, err)
	}
	if file, err := pages.Audio("w", "Дом"); err != nil || file.ContentType != "audio/mp4" {
		t.Errorf("audio: %+v %v", file, err)
	}
	if _, err := pages.Audio("x", "дом"); !errors.Is(err, ErrNotFound) {
		t.Errorf("unknown kind: %v", err)
	}
	recorded, missing, err := pages.Missing()
	if err != nil || recorded != 1 || len(missing) != len(f.page.Clips())-1 {
		t.Errorf("missing: %d %d %v", recorded, len(missing), err)
	}
}
