package content

import (
	"errors"
	"fmt"
	"strings"

	"github.com/artim-ctrl/domashka/internal/domain/clip"
)

const (
	russianVowels = "аоуыиэеёюя"
	stressMark    = "+"
)

func validateSpeech(clips []clip.Clip, speech map[string]string) error {
	var errs []error
	used := make(map[string]bool, len(clips))
	for i := range clips {
		c := clips[i]
		used[c.ID()] = true
		text, ok := speech[c.ID()]
		switch {
		case !ok:
			errs = append(errs, fmt.Errorf("no speech text for %s", c.ID()))
		case c.Kind == clip.Word && strings.ReplaceAll(text, stressMark, "") != c.Key:
			errs = append(errs, fmt.Errorf("%s: %q is not the same word", c.ID(), text))
		case c.Kind == clip.Sound && !strings.ContainsAny(c.Key, russianVowels) && text == c.Key:
			errs = append(errs, fmt.Errorf("%s: a lone consonant is read by synthesizers as the letter name", c.ID()))
		}
	}
	for id := range speech {
		if !used[id] {
			errs = append(errs, fmt.Errorf("%s is not on the page", id))
		}
	}
	return errors.Join(errs...)
}
