package voiceover

import (
	"errors"
	"fmt"
	"slices"
)

type Engine string

const (
	EngineSilero Engine = "silero"
	EngineSay    Engine = "say"
)

type Voice struct {
	ID      string
	Label   string
	License string
	Engine  Engine
	Model   string
	Speaker string
}

type Catalog struct {
	voices []Voice
}

func NewCatalog(voices []Voice) (*Catalog, error) {
	if len(voices) == 0 {
		return nil, errors.New("voice catalog is empty")
	}
	seen := make(map[string]bool, len(voices))
	for i := range voices {
		switch voices[i].Engine {
		case EngineSilero, EngineSay:
		default:
			return nil, fmt.Errorf("voice %s has unknown engine %q", voices[i].ID, voices[i].Engine)
		}
		if seen[voices[i].ID] {
			return nil, fmt.Errorf("voice %s is listed twice", voices[i].ID)
		}
		seen[voices[i].ID] = true
	}
	return &Catalog{voices: slices.Clone(voices)}, nil
}

func (c *Catalog) All() []Voice { return slices.Clone(c.voices) }

func (c *Catalog) ByID(id string) (Voice, bool) {
	i := slices.IndexFunc(c.voices, func(v Voice) bool { return v.ID == id })
	if i < 0 {
		return Voice{}, false
	}
	return c.voices[i], true
}

type Choice struct {
	Voice string `json:"voice"`
	Text  string `json:"text"`
}

type Candidate struct {
	Voice   string
	Version int64
}

type Recording struct {
	Path    string
	Version int64
}
