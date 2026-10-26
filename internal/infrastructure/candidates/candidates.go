package candidates

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/artim-ctrl/domashka/internal/domain/clip"
	"github.com/artim-ctrl/domashka/internal/domain/voiceover"
)

type Options struct {
	Dir      string
	KindDirs clip.Dirs
	TextFile string
	Ext      string
}

type Store struct {
	Options
}

func New(opts Options) *Store {
	return &Store{Options: opts}
}

func (s *Store) Path(c clip.Clip, voiceID string) string {
	return filepath.Join(s.clipDir(c), voiceID+s.Ext)
}

func (s *Store) Get(c clip.Clip, voiceID string) (voiceover.Candidate, bool, error) {
	info, err := os.Stat(s.Path(c, voiceID))
	if errors.Is(err, os.ErrNotExist) {
		return voiceover.Candidate{}, false, nil
	}
	if err != nil {
		return voiceover.Candidate{}, false, err
	}
	return voiceover.Candidate{Voice: voiceID, Version: info.ModTime().UnixMilli()}, true, nil
}

func (s *Store) List(c clip.Clip, voices []voiceover.Voice) ([]voiceover.Candidate, error) {
	cands := []voiceover.Candidate{}
	for i := range voices {
		cand, ok, err := s.Get(c, voices[i].ID)
		if err != nil {
			return nil, err
		}
		if ok {
			cands = append(cands, cand)
		}
	}
	return cands, nil
}

func (s *Store) Text(c clip.Clip) (string, error) {
	b, err := os.ReadFile(s.textPath(c))
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	return string(b), err
}

func (s *Store) SaveText(c clip.Clip, text string) error {
	if err := os.MkdirAll(s.clipDir(c), 0o755); err != nil {
		return err
	}
	return os.WriteFile(s.textPath(c), []byte(text), 0o644)
}

func (s *Store) clipDir(c clip.Clip) string {
	return filepath.Join(s.Dir, s.KindDirs.For(c.Kind), c.Key)
}

func (s *Store) textPath(c clip.Clip) string { return filepath.Join(s.clipDir(c), s.TextFile) }
