package config

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type validator struct {
	errs []error
}

func (v *validator) required(name, value string) {
	if strings.TrimSpace(value) == "" {
		v.errs = append(v.errs, fmt.Errorf("%s is required", name))
	}
}

func (v *validator) positive(name string, value int64) {
	if value <= 0 {
		v.errs = append(v.errs, fmt.Errorf("%s must be positive, got %d", name, value))
	}
}

func (v *validator) duration(name string, value time.Duration) {
	if value <= 0 {
		v.errs = append(v.errs, fmt.Errorf("%s must be a positive duration, got %s", name, value))
	}
}

func (v *validator) ext(name, value string) {
	v.required(name, value)
	if value != "" && !strings.HasPrefix(value, ".") {
		v.errs = append(v.errs, fmt.Errorf("%s must start with a dot, got %q", name, value))
	}
}

func (v *validator) kindDirs(name string, dirs KindDirs) {
	v.required(name+".sounds", dirs.Sounds)
	v.required(name+".words", dirs.Words)
	if dirs.Sounds != "" && dirs.Sounds == dirs.Words {
		v.errs = append(v.errs, fmt.Errorf("%s.sounds and %s.words must differ", name, name))
	}
}

func (c *Config) validate() error {
	v := &validator{}

	v.required("server.addr", c.Server.Addr)
	if c.Server.BasePath != "" && !strings.HasPrefix(c.Server.BasePath, "/") {
		v.errs = append(v.errs, fmt.Errorf("server.base_path must start with /, got %q", c.Server.BasePath))
	}
	v.positive("server.read_buffer_size", int64(c.Server.ReadBufferSize))
	v.positive("server.body_limit", int64(c.Server.BodyLimit))
	t := c.Server.Timeouts
	v.duration("server.timeouts.startup", t.Startup)
	v.duration("server.timeouts.read", t.Read)
	v.duration("server.timeouts.write", t.Write)
	v.duration("server.timeouts.admin_write", t.AdminWrite)
	v.duration("server.timeouts.idle", t.Idle)
	v.duration("server.timeouts.shutdown", t.Shutdown)
	v.duration("server.timeouts.healthcheck", t.Healthcheck)

	v.required("content.page", c.Content.Page)
	v.required("content.speech", c.Content.Speech)

	v.required("library.dir", c.Library.Dir)
	v.kindDirs("library.kind_dirs", c.Library.KindDirs)
	v.required("library.choices_file", c.Library.ChoicesFile)
	v.ext("library.recording_ext", c.Library.RecordingExt)
	v.formats(c.Library.Formats, c.Library.RecordingExt, c.Candidates.Ext)

	v.required("candidates.dir", c.Candidates.Dir)
	v.kindDirs("candidates.kind_dirs", c.Candidates.KindDirs)
	v.required("candidates.text_file", c.Candidates.TextFile)
	v.ext("candidates.ext", c.Candidates.Ext)

	v.positive("admin.max_speech_text_runes", int64(c.Admin.MaxSpeechTextRunes))

	if len(c.TTS.Silero.Command) == 0 {
		v.errs = append(v.errs, errors.New("tts.silero.command is required"))
	}
	v.required("tts.silero.script", c.TTS.Silero.Script)
	v.positive("tts.say.rate", int64(c.TTS.Say.Rate))
	v.voices(c.TTS.Voices)

	v.positive("encoding.wav_sample_rate", int64(c.Encoding.WAVSampleRate))
	v.positive("encoding.aac_bitrate", int64(c.Encoding.AACBitrate))

	return errors.Join(v.errs...)
}

func (v *validator) voices(voices []Voice) {
	if len(voices) == 0 {
		v.errs = append(v.errs, errors.New("tts.voices must list at least one voice"))
	}
	seen := make(map[string]bool, len(voices))
	for i := range voices {
		name := fmt.Sprintf("tts.voices[%d]", i)
		vc := voices[i]
		v.required(name+".id", vc.ID)
		v.required(name+".label", vc.Label)
		v.required(name+".license", vc.License)
		v.required(name+".model", vc.Model)
		switch vc.Engine {
		case "silero":
			v.required(name+".speaker", vc.Speaker)
		case "say":
		default:
			v.errs = append(v.errs, fmt.Errorf("%s.engine must be silero or say, got %q", name, vc.Engine))
		}
		if seen[vc.ID] {
			v.errs = append(v.errs, fmt.Errorf("%s.id %q is duplicated", name, vc.ID))
		}
		seen[vc.ID] = true
	}
}

func (v *validator) formats(formats []Format, required ...string) {
	if len(formats) == 0 {
		v.errs = append(v.errs, errors.New("library.formats must list at least one format"))
	}
	seen := make(map[string]bool, len(formats))
	for i := range formats {
		name := fmt.Sprintf("library.formats[%d]", i)
		v.ext(name+".ext", formats[i].Ext)
		v.required(name+".content_type", formats[i].ContentType)
		if formats[i].Ext != strings.ToLower(formats[i].Ext) {
			v.errs = append(v.errs, fmt.Errorf("%s.ext must be lowercase, got %q", name, formats[i].Ext))
		}
		if seen[formats[i].Ext] {
			v.errs = append(v.errs, fmt.Errorf("%s.ext %q is duplicated", name, formats[i].Ext))
		}
		seen[formats[i].Ext] = true
	}
	for i := range required {
		if required[i] != "" && !seen[required[i]] {
			v.errs = append(v.errs, fmt.Errorf("library.formats must include %q", required[i]))
		}
	}
}
