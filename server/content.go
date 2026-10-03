package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"regexp"
)

// Текст страницы. В строках можно писать **жирный**, *курсив* и метки звука:
// {{s:ш}} — звук, {{w:дом}} — слово, {{w:молоко|молоко́}} — ключ записи|что показать.
//
//go:embed content/page.json
var pageJSON []byte

type Page struct {
	Title    string    `json:"title"`
	Sections []Section `json:"sections"`
}

type Section struct {
	ID     string  `json:"id"`
	Title  string  `json:"title,omitempty"`
	Blocks []Block `json:"blocks"`
}

type Block struct {
	Type    string      `json:"type"` // paragraph | callout | list | table | alphabet
	Text    string      `json:"text,omitempty"`
	Style   string      `json:"style,omitempty"`
	Items   []string    `json:"items,omitempty"`
	Columns []string    `json:"columns,omitempty"`
	Rows    [][]string  `json:"rows,omitempty"`
	Letters []LetterRow `json:"letters,omitempty"`
}

type LetterRow struct {
	Letter  string `json:"letter"`
	Sound   string `json:"sound,omitempty"` // пусто у ъ и ь — своего звука нет
	IPA     string `json:"ipa"`
	ES      string `json:"es"`
	EN      string `json:"en"`
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
	for _, s := range p.Sections {
		for _, b := range s.Blocks {
			if !blockTypes[b.Type] {
				return nil, fmt.Errorf("content/page.json: раздел %q: неизвестный блок %q", s.ID, b.Type)
			}
		}
	}
	return &p, nil
}

var tokenRe = regexp.MustCompile(`\{\{([sw]):([^}|]+)(?:\|[^}]*)?\}\}`)

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
	scan := func(text string) {
		for _, m := range tokenRe.FindAllStringSubmatch(text, -1) {
			add(Clip{Kind: Kind(m[1]), Key: m[2]})
		}
	}
	for _, s := range p.Sections {
		for _, b := range s.Blocks {
			scan(b.Text)
			for _, it := range b.Items {
				scan(it)
			}
			for _, row := range b.Rows {
				for _, cell := range row {
					scan(cell)
				}
			}
			for _, l := range b.Letters {
				add(Clip{Kind: KindSound, Key: l.Sound})
				add(Clip{Kind: KindWord, Key: l.Example})
				scan(l.ES)
				scan(l.EN)
			}
		}
	}
	return clips
}
