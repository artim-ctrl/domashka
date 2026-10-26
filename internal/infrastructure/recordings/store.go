package recordings

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/artim-ctrl/domashka/internal/domain/clip"
	"github.com/artim-ctrl/domashka/internal/domain/voiceover"
)

type Encoder interface {
	ToM4A(ctx context.Context, in, out string) error
}

type Options struct {
	Dir          string
	KindDirs     clip.Dirs
	ChoicesFile  string
	RecordingExt string
	Formats      []Format
	Encoder      Encoder
}

type Store struct {
	Options

	mu     sync.Mutex
	frozen map[clip.Clip]voiceover.Recording
}

func New(opts Options) *Store {
	return &Store{Options: opts}
}

func (l *Store) Freeze() error {
	all, err := l.scanDir()
	if err != nil {
		return err
	}
	l.frozen = all
	return nil
}

func (l *Store) Scan() (map[clip.Clip]voiceover.Recording, error) {
	if l.frozen != nil {
		return l.frozen, nil
	}
	return l.scanDir()
}

func (l *Store) Lookup(c clip.Clip) (voiceover.Recording, bool, error) {
	all, err := l.Scan()
	if err != nil {
		return voiceover.Recording{}, false, err
	}
	rec, ok := all[c.Normalized()]
	return rec, ok, nil
}

func (l *Store) ExpectedPath(c clip.Clip) string {
	return filepath.Join(l.kindDir(c.Kind), c.Key+l.RecordingExt)
}

func (l *Store) kindDir(k clip.Kind) string { return filepath.Join(l.Dir, l.KindDirs.For(k)) }

func (l *Store) scanDir() (map[clip.Clip]voiceover.Recording, error) {
	found := map[clip.Clip]voiceover.Recording{}
	kinds := []clip.Kind{clip.Sound, clip.Word}
	for i := range kinds {
		if err := l.scanKind(kinds[i], found); err != nil {
			return nil, err
		}
	}
	return found, nil
}

func (l *Store) scanKind(kind clip.Kind, found map[clip.Clip]voiceover.Recording) error {
	dir := l.kindDir(kind)
	entries, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	rank := map[clip.Clip]int{}
	for i := range entries {
		e := entries[i]
		r := l.formatRank(e.Name())
		if e.IsDir() || r < 0 {
			continue
		}
		c := clip.Clip{Kind: kind, Key: keyOf(e.Name())}.Normalized()
		if prev, ok := rank[c]; ok && prev <= r {
			continue
		}
		info, err := e.Info()
		if err != nil {
			return err
		}
		rank[c] = r
		found[c] = voiceover.Recording{Path: filepath.Join(dir, e.Name()), Version: info.ModTime().UnixMilli()}
	}
	return nil
}

func keyOf(fileName string) string {
	return strings.TrimSuffix(fileName, filepath.Ext(fileName))
}
