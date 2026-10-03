package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
)

// Текст страницы. В строках можно писать **жирный**, *курсив* и метки звука:
// {{s:ш}} — звук, {{w:дом}} — слово, {{w:молоко|молоко́}} — ключ записи|что показать.
//
//go:embed content/page.json
var pageJSON []byte

// Языки интерфейса; первый — запасной.
var langs = []string{"es", "en", "ru"}

// Text — строка на всех языках. В JSON это либо объект {"es": …, "en": …, "ru": …},
// либо просто строка, если она одинакова на всех языках («{{w:база}} — {{w:ваза}}»).
type Text struct {
	same string
	by   map[string]string
}

func (t *Text) UnmarshalJSON(b []byte) error {
	if len(b) > 0 && b[0] == '"' {
		return json.Unmarshal(b, &t.same)
	}
	return json.Unmarshal(b, &t.by)
}

func (t Text) MarshalJSON() ([]byte, error) {
	if t.by == nil {
		return json.Marshal(t.same)
	}
	return json.Marshal(t.by)
}

// Variants — сам текст на каждом языке (одна строка, если он общий).
func (t Text) Variants() []string {
	if t.by == nil {
		return []string{t.same}
	}
	out := make([]string, 0, len(t.by))
	for _, l := range langs {
		out = append(out, t.by[l])
	}
	return out
}

func (t Text) check() error {
	if t.by == nil {
		return nil
	}
	for l := range t.by {
		if !slices.Contains(langs, l) {
			return fmt.Errorf("неизвестный язык %q", l)
		}
	}
	want := tokens(t.by[langs[0]])
	for _, l := range langs {
		s, ok := t.by[l]
		if !ok || strings.TrimSpace(s) == "" {
			return fmt.Errorf("нет перевода на %q: %q", l, t.by[langs[0]])
		}
		if got := tokens(s); !slices.Equal(got, want) {
			return fmt.Errorf("в переводе на %q другие кнопки: %v вместо %v (%q)", l, got, want, s)
		}
	}
	return nil
}

type Page struct {
	Title    Text      `json:"title"`
	Sections []Section `json:"sections"`
}

type Section struct {
	ID     string  `json:"id"`
	Title  *Text   `json:"title,omitempty"`
	Blocks []Block `json:"blocks"`
}

type Block struct {
	Type    string      `json:"type"` // paragraph | callout | list | table | alphabet
	Text    *Text       `json:"text,omitempty"`
	Style   string      `json:"style,omitempty"`
	Items   []Text      `json:"items,omitempty"`
	Columns []Text      `json:"columns,omitempty"`
	Rows    [][]Text    `json:"rows,omitempty"`
	Letters []LetterRow `json:"letters,omitempty"`
}

type LetterRow struct {
	Letter  string `json:"letter"`
	Sound   string `json:"sound,omitempty"` // пусто у ъ и ь — своего звука нет
	IPA     Text   `json:"ipa"`
	Spanish Text   `json:"spanish"` // на что похоже в испанском
	English Text   `json:"english"` // на что похоже в английском
	Example string `json:"example"`
	Focus   bool   `json:"focus,omitempty"`
}

var blockTypes = map[string]bool{"paragraph": true, "callout": true, "list": true, "table": true, "alphabet": true}

func loadPage() (*Page, error) {
	dec := json.NewDecoder(bytes.NewReader(pageJSON))
	dec.DisallowUnknownFields()
	var p Page
	if err := dec.Decode(&p); err != nil {
		return nil, fmt.Errorf("content/page.json: %w", err)
	}
	var errs []error
	for _, s := range p.Sections {
		for _, b := range s.Blocks {
			if !blockTypes[b.Type] {
				errs = append(errs, fmt.Errorf("раздел %q: неизвестный блок %q", s.ID, b.Type))
			}
		}
	}
	p.eachText(func(t Text) {
		if err := t.check(); err != nil {
			errs = append(errs, err)
		}
	})
	if err := errors.Join(errs...); err != nil {
		return nil, fmt.Errorf("content/page.json:\n%w", err)
	}
	return &p, nil
}

func (p *Page) eachText(fn func(Text)) {
	opt := func(t *Text) {
		if t != nil {
			fn(*t)
		}
	}
	fn(p.Title)
	for _, s := range p.Sections {
		opt(s.Title)
		for _, b := range s.Blocks {
			opt(b.Text)
			for _, t := range b.Items {
				fn(t)
			}
			for _, t := range b.Columns {
				fn(t)
			}
			for _, row := range b.Rows {
				for _, t := range row {
					fn(t)
				}
			}
			for _, l := range b.Letters {
				fn(l.IPA)
				fn(l.Spanish)
				fn(l.English)
			}
		}
	}
}

var tokenRe = regexp.MustCompile(`\{\{([sw]):([^}|]+)(?:\|[^}]*)?\}\}`)

// tokens — кнопки в строке, «s:ш», «w:дом», по порядку.
func tokens(s string) []string {
	var out []string
	for _, m := range tokenRe.FindAllStringSubmatch(s, -1) {
		out = append(out, m[1]+":"+m[2])
	}
	return out
}

// Clips — все записи, которые упоминает страница, в порядке появления, без повторов.
func (p *Page) Clips() []Clip {
	var clips []Clip
	seen := map[Clip]bool{}
	add := func(c Clip) {
		if c.Key != "" && !seen[c] {
			seen[c] = true
			clips = append(clips, c)
		}
	}
	for _, s := range p.Sections {
		for _, b := range s.Blocks {
			for _, l := range b.Letters {
				add(Clip{Kind: KindSound, Key: l.Sound})
				add(Clip{Kind: KindWord, Key: l.Example})
			}
		}
	}
	// Переводы проверены на одинаковые кнопки, так что хватает испанского.
	p.eachText(func(t Text) {
		for _, tok := range tokens(t.Variants()[0]) {
			kind, key, _ := strings.Cut(tok, ":")
			add(Clip{Kind: Kind(kind), Key: key})
		}
	})
	return clips
}
