package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"go.yaml.in/yaml/v3"
)

type Config struct {
	Server     Server     `yaml:"server"`
	Content    Content    `yaml:"content"`
	Library    Library    `yaml:"library"`
	Candidates Candidates `yaml:"candidates"`
	Admin      Admin      `yaml:"admin"`
	TTS        TTS        `yaml:"tts"`
	Encoding   Encoding   `yaml:"encoding"`
}

type Server struct {
	Addr           string   `yaml:"addr"`
	BasePath       string   `yaml:"base_path"`
	StaticDir      string   `yaml:"static_dir"`
	ReadBufferSize int      `yaml:"read_buffer_size"`
	BodyLimit      int      `yaml:"body_limit"`
	Timeouts       Timeouts `yaml:"timeouts"`
}

type Timeouts struct {
	Startup     time.Duration `yaml:"startup"`
	Read        time.Duration `yaml:"read"`
	Write       time.Duration `yaml:"write"`
	AdminWrite  time.Duration `yaml:"admin_write"`
	Idle        time.Duration `yaml:"idle"`
	Shutdown    time.Duration `yaml:"shutdown"`
	Healthcheck time.Duration `yaml:"healthcheck"`
}

type Content struct {
	Page   string `yaml:"page"`
	Speech string `yaml:"speech"`
}

type KindDirs struct {
	Sounds string `yaml:"sounds"`
	Words  string `yaml:"words"`
}

type Library struct {
	Dir          string   `yaml:"dir"`
	KindDirs     KindDirs `yaml:"kind_dirs"`
	ChoicesFile  string   `yaml:"choices_file"`
	RecordingExt string   `yaml:"recording_ext"`
	Formats      []Format `yaml:"formats"`
}

type Format struct {
	Ext         string `yaml:"ext"`
	ContentType string `yaml:"content_type"`
}

type Candidates struct {
	Dir      string   `yaml:"dir"`
	KindDirs KindDirs `yaml:"kind_dirs"`
	TextFile string   `yaml:"text_file"`
	Ext      string   `yaml:"ext"`
}

type Admin struct {
	MaxSpeechTextRunes int `yaml:"max_speech_text_runes"`
}

type TTS struct {
	Silero Silero  `yaml:"silero"`
	Say    Say     `yaml:"say"`
	Voices []Voice `yaml:"voices"`
}

type Silero struct {
	Command []string `yaml:"command"`
	Script  string   `yaml:"script"`
	Env     []string `yaml:"env"`
}

type Say struct {
	Rate int `yaml:"rate"`
}

type Voice struct {
	ID      string `yaml:"id"`
	Label   string `yaml:"label"`
	License string `yaml:"license"`
	Engine  string `yaml:"engine"`
	Model   string `yaml:"model"`
	Speaker string `yaml:"speaker"`
}

type Encoding struct {
	WAVSampleRate int `yaml:"wav_sample_rate"`
	AACBitrate    int `yaml:"aac_bitrate"`
}

func Load(paths []string) (Config, error) {
	if len(paths) == 0 {
		return Config{}, errors.New("no config files given")
	}
	var cfg Config
	for i := range paths {
		if err := decodeFile(paths[i], &cfg); err != nil {
			return Config{}, err
		}
	}
	if err := cfg.validate(); err != nil {
		return Config{}, fmt.Errorf("config %v:\n%w", paths, err)
	}
	return cfg, nil
}

func decodeFile(path string, cfg *Config) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}
	dec := yaml.NewDecoder(bytes.NewReader(b))
	dec.KnownFields(true)
	if err := dec.Decode(cfg); err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("config %s: %w", path, err)
	}
	return nil
}
