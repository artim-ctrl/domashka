package tts

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/artim-ctrl/domashka/internal/domain/voiceover"
)

type WAVConverter interface {
	ToWAV(ctx context.Context, in, out string) error
}

type Options struct {
	SileroCommand []string
	SileroScript  string
	SileroEnv     []string
	SayRate       int
	Converter     WAVConverter
}

type Synthesizer struct {
	silero    *sileroWorker
	sayRate   int
	converter WAVConverter
}

func NewSynthesizer(opts Options) *Synthesizer {
	return &Synthesizer{
		silero: &sileroWorker{
			command: append(append([]string{}, opts.SileroCommand...), opts.SileroScript),
			env:     opts.SileroEnv,
		},
		sayRate:   opts.SayRate,
		converter: opts.Converter,
	}
}

func (s *Synthesizer) Synthesize(ctx context.Context, v voiceover.Voice, text, outWAV string) error {
	switch v.Engine {
	case voiceover.EngineSay:
		return s.say(ctx, v.Model, strings.ReplaceAll(text, "+", ""), outWAV)
	case voiceover.EngineSilero:
		return s.silero.synthesize(v, text, outWAV)
	default:
		return fmt.Errorf("voice %s has unknown engine %q", v.ID, v.Engine)
	}
}

func (s *Synthesizer) Close() { s.silero.close() }

func (s *Synthesizer) say(ctx context.Context, voice, text, outWAV string) error {
	aiff := outWAV + ".aiff"
	defer os.Remove(aiff)
	cmd := exec.CommandContext(ctx, "say", "-v", voice, "-r", strconv.Itoa(s.sayRate), "-o", aiff, "--", text)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("say: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return s.converter.ToWAV(ctx, aiff, outWAV)
}
