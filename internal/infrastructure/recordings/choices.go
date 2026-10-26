package recordings

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/artim-ctrl/domashka/internal/domain/clip"
	"github.com/artim-ctrl/domashka/internal/domain/voiceover"
)

func (l *Store) choicesPath() string { return filepath.Join(l.Dir, l.ChoicesFile) }

func (l *Store) Choices() (map[string]voiceover.Choice, error) {
	choices := map[string]voiceover.Choice{}
	b, err := os.ReadFile(l.choicesPath())
	if errors.Is(err, os.ErrNotExist) {
		return choices, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(b, &choices); err != nil {
		return nil, err
	}
	return choices, nil
}

func (l *Store) Choose(ctx context.Context, c clip.Clip, wav string, choice voiceover.Choice) (voiceover.Recording, error) {
	if l.Encoder == nil {
		return voiceover.Recording{}, errors.New("library has no encoder: choosing recordings needs the admin mode")
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	dir := l.kindDir(c.Kind)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return voiceover.Recording{}, err
	}
	final := filepath.Join(dir, c.Key+l.RecordingExt)
	if err := writeAtomically(final, func(tmp string) error { return l.Encoder.ToM4A(ctx, wav, tmp) }); err != nil {
		return voiceover.Recording{}, err
	}
	if err := l.removeOtherFormats(dir, c.Key, final); err != nil {
		return voiceover.Recording{}, err
	}
	if err := l.saveChoice(c, choice); err != nil {
		return voiceover.Recording{}, err
	}

	info, err := os.Stat(final)
	if err != nil {
		return voiceover.Recording{}, err
	}
	return voiceover.Recording{Path: final, Version: info.ModTime().UnixMilli()}, nil
}

func (l *Store) saveChoice(c clip.Clip, choice voiceover.Choice) error {
	choices, err := l.Choices()
	if err != nil {
		return err
	}
	choices[c.ID()] = choice
	b, err := json.MarshalIndent(choices, "", "  ")
	if err != nil {
		return err
	}
	return writeAtomically(l.choicesPath(), func(tmp string) error {
		return os.WriteFile(tmp, append(b, '\n'), 0o644)
	})
}

func writeAtomically(path string, write func(tmp string) error) error {
	tmp := filepath.Join(filepath.Dir(path), "."+filepath.Base(path)+".tmp")
	if err := write(tmp); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, path)
}

func (l *Store) removeOtherFormats(dir, key, keep string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	want := clip.NormalizeKey(key)
	var errs []error
	for i := range entries {
		name := entries[i].Name()
		path := filepath.Join(dir, name)
		if path == keep || entries[i].IsDir() || l.formatRank(name) < 0 {
			continue
		}
		if clip.NormalizeKey(keyOf(name)) == want {
			errs = append(errs, os.Remove(path))
		}
	}
	return errors.Join(errs...)
}
