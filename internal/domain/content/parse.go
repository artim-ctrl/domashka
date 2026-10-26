package content

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

func Parse(pageData, speechData []byte) (*Page, error) {
	dec := json.NewDecoder(bytes.NewReader(pageData))
	dec.DisallowUnknownFields()
	var p Page
	if err := dec.Decode(&p); err != nil {
		return nil, fmt.Errorf("page: %w", err)
	}
	if err := p.validate(); err != nil {
		return nil, fmt.Errorf("page:\n%w", err)
	}
	p.clips = p.collectClips()

	if err := json.Unmarshal(speechData, &p.speech); err != nil {
		return nil, fmt.Errorf("speech: %w", err)
	}
	if err := validateSpeech(p.clips, p.speech); err != nil {
		return nil, fmt.Errorf("speech:\n%w", err)
	}
	return &p, nil
}

func (p *Page) validate() error {
	var errs []error
	for i := range p.Sections {
		s := &p.Sections[i]
		for j := range s.Blocks {
			if !s.Blocks[j].Type.valid() {
				errs = append(errs, fmt.Errorf("section %q: unknown block type %q", s.ID, s.Blocks[j].Type))
			}
		}
	}
	p.eachText(func(t Text) {
		if err := t.validate(); err != nil {
			errs = append(errs, err)
		}
	})
	return errors.Join(errs...)
}
