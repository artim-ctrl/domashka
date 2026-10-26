package bootstrap

import (
	"context"

	"go.uber.org/fx"

	"github.com/artim-ctrl/domashka/internal/application"
	"github.com/artim-ctrl/domashka/internal/domain/clip"
	"github.com/artim-ctrl/domashka/internal/domain/content"
	"github.com/artim-ctrl/domashka/internal/domain/voiceover"
	"github.com/artim-ctrl/domashka/internal/infrastructure/audioconv"
	"github.com/artim-ctrl/domashka/internal/infrastructure/candidates"
	"github.com/artim-ctrl/domashka/internal/infrastructure/config"
	"github.com/artim-ctrl/domashka/internal/infrastructure/contentfile"
	"github.com/artim-ctrl/domashka/internal/infrastructure/recordings"
	"github.com/artim-ctrl/domashka/internal/infrastructure/tts"
)

type Runtime struct {
	Admin bool
}

func kindDirs(d config.KindDirs) clip.Dirs {
	return clip.Dirs{Sounds: d.Sounds, Words: d.Words}
}

func newConverter(cfg config.Config) audioconv.Converter {
	return audioconv.Converter{WAVSampleRate: cfg.Encoding.WAVSampleRate, AACBitrate: cfg.Encoding.AACBitrate}
}

func loadPage(cfg config.Config) (*content.Page, error) {
	return contentfile.Load(cfg.Content.Page, cfg.Content.Speech)
}

func newRecordingStore(cfg config.Config, rt Runtime, conv audioconv.Converter) (application.RecordingStore, error) {
	formats := make([]recordings.Format, 0, len(cfg.Library.Formats))
	for i := range cfg.Library.Formats {
		formats = append(formats, recordings.Format{Ext: cfg.Library.Formats[i].Ext, ContentType: cfg.Library.Formats[i].ContentType})
	}
	opts := recordings.Options{
		Dir:          cfg.Library.Dir,
		KindDirs:     kindDirs(cfg.Library.KindDirs),
		ChoicesFile:  cfg.Library.ChoicesFile,
		RecordingExt: cfg.Library.RecordingExt,
		Formats:      formats,
	}
	if rt.Admin {
		opts.Encoder = conv
	}
	store := recordings.New(opts)
	if !rt.Admin {
		if err := store.Freeze(); err != nil {
			return nil, err
		}
	}
	return store, nil
}

func newCatalog(cfg config.Config) (*voiceover.Catalog, error) {
	voices := make([]voiceover.Voice, 0, len(cfg.TTS.Voices))
	for i := range cfg.TTS.Voices {
		v := cfg.TTS.Voices[i]
		voices = append(voices, voiceover.Voice{
			ID:      v.ID,
			Label:   v.Label,
			License: v.License,
			Engine:  voiceover.Engine(v.Engine),
			Model:   v.Model,
			Speaker: v.Speaker,
		})
	}
	return voiceover.NewCatalog(voices)
}

func newSynthesizer(lc fx.Lifecycle, cfg config.Config, conv audioconv.Converter) application.Synthesizer {
	synth := tts.NewSynthesizer(tts.Options{
		SileroCommand: cfg.TTS.Silero.Command,
		SileroScript:  cfg.TTS.Silero.Script,
		SileroEnv:     cfg.TTS.Silero.Env,
		SayRate:       cfg.TTS.Say.Rate,
		Converter:     conv,
	})
	lc.Append(fx.StopHook(func(context.Context) error {
		synth.Close()
		return nil
	}))
	return synth
}

func newCandidateStore(cfg config.Config) application.CandidateStore {
	return candidates.New(candidates.Options{
		Dir:      cfg.Candidates.Dir,
		KindDirs: kindDirs(cfg.Candidates.KindDirs),
		TextFile: cfg.Candidates.TextFile,
		Ext:      cfg.Candidates.Ext,
	})
}

func newVoiceoverService(
	page *content.Page,
	voices *voiceover.Catalog,
	recs application.RecordingStore,
	cands application.CandidateStore,
	synth application.Synthesizer,
	cfg config.Config,
) *application.VoiceoverService {
	return application.NewVoiceoverService(page, voices, recs, cands, synth, application.VoiceoverOptions{
		MaxTextRunes: cfg.Admin.MaxSpeechTextRunes,
	})
}
