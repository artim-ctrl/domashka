package application

import (
	"fmt"

	"github.com/artim-ctrl/domashka/internal/domain/clip"
	"github.com/artim-ctrl/domashka/internal/domain/content"
	"github.com/artim-ctrl/domashka/internal/domain/voiceover"
)

type PageView struct {
	Page       *content.Page
	Clips      []clip.Clip
	Recordings map[clip.Clip]voiceover.Recording
}

type AudioFile struct {
	Recording   voiceover.Recording
	ContentType string
}

type PageService struct {
	page       *content.Page
	recordings RecordingStore
}

func NewPageService(page *content.Page, recordings RecordingStore) *PageService {
	return &PageService{page: page, recordings: recordings}
}

func (s *PageService) View() (PageView, error) {
	have, err := s.recordings.Scan()
	if err != nil {
		return PageView{}, err
	}
	clips := s.page.Clips()
	view := PageView{Page: s.page, Clips: clips, Recordings: make(map[clip.Clip]voiceover.Recording, len(clips))}
	for i := range clips {
		if rec, ok := have[clips[i].Normalized()]; ok {
			view.Recordings[clips[i]] = rec
		}
	}
	return view, nil
}

func (s *PageService) Audio(kind, key string) (AudioFile, error) {
	k, ok := clip.ParseKind(kind)
	if !ok {
		return AudioFile{}, fmt.Errorf("clip kind %q: %w", kind, ErrNotFound)
	}
	rec, ok, err := s.recordings.Lookup(clip.Clip{Kind: k, Key: key})
	if err != nil {
		return AudioFile{}, err
	}
	if !ok {
		return AudioFile{}, fmt.Errorf("recording %s:%s: %w", kind, key, ErrNotFound)
	}
	return AudioFile{Recording: rec, ContentType: s.recordings.ContentType(rec.Path)}, nil
}

func (s *PageService) Missing() (recorded int, missing []string, err error) {
	have, err := s.recordings.Scan()
	if err != nil {
		return 0, nil, err
	}
	clips := s.page.Clips()
	for i := range clips {
		if _, ok := have[clips[i].Normalized()]; !ok {
			missing = append(missing, s.recordings.ExpectedPath(clips[i]))
		}
	}
	return len(clips) - len(missing), missing, nil
}
