package content

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
)

var languages = []string{"es", "en", "ru"}

type Text struct {
	same   string
	byLang map[string]string
}

func (t *Text) UnmarshalJSON(b []byte) error {
	if len(b) > 0 && b[0] == '"' {
		return json.Unmarshal(b, &t.same)
	}
	return json.Unmarshal(b, &t.byLang)
}

func (t Text) MarshalJSON() ([]byte, error) {
	if t.byLang == nil {
		return json.Marshal(t.same)
	}
	return json.Marshal(t.byLang)
}

func (t Text) primary() string {
	if t.byLang == nil {
		return t.same
	}
	return t.byLang[languages[0]]
}

func (t Text) validate() error {
	if t.byLang == nil {
		return nil
	}
	for lang := range t.byLang {
		if !slices.Contains(languages, lang) {
			return fmt.Errorf("unknown language %q", lang)
		}
	}
	want := tokens(t.primary())
	for i := range languages {
		lang := languages[i]
		s, ok := t.byLang[lang]
		if !ok || strings.TrimSpace(s) == "" {
			return fmt.Errorf("missing %q translation of %q", lang, t.primary())
		}
		if got := tokens(s); !slices.Equal(got, want) {
			return fmt.Errorf("%q translation has different clips: %v instead of %v (%q)", lang, got, want, s)
		}
	}
	return nil
}
