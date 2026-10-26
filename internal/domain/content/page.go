package content

import (
	"github.com/artim-ctrl/domashka/internal/domain/clip"
)

type Page struct {
	Title    Text      `json:"title"`
	Sections []Section `json:"sections"`

	speech map[string]string
	clips  []clip.Clip
}

type Section struct {
	ID     string  `json:"id"`
	Title  *Text   `json:"title,omitempty"`
	Blocks []Block `json:"blocks"`
}

type BlockType string

const (
	BlockParagraph BlockType = "paragraph"
	BlockCallout   BlockType = "callout"
	BlockList      BlockType = "list"
	BlockTable     BlockType = "table"
	BlockAlphabet  BlockType = "alphabet"
)

func (t BlockType) valid() bool {
	switch t {
	case BlockParagraph, BlockCallout, BlockList, BlockTable, BlockAlphabet:
		return true
	default:
		return false
	}
}

type Block struct {
	Type    BlockType   `json:"type"`
	Text    *Text       `json:"text,omitempty"`
	Style   string      `json:"style,omitempty"`
	Items   []Text      `json:"items,omitempty"`
	Columns []Text      `json:"columns,omitempty"`
	Rows    [][]Text    `json:"rows,omitempty"`
	Letters []LetterRow `json:"letters,omitempty"`
}

type LetterRow struct {
	Letter  string `json:"letter"`
	Sound   string `json:"sound,omitempty"`
	IPA     Text   `json:"ipa"`
	Spanish Text   `json:"spanish"`
	English Text   `json:"english"`
	Example string `json:"example"`
	Focus   bool   `json:"focus,omitempty"`
}

func (p *Page) Clips() []clip.Clip { return p.clips }

func (p *Page) Spoken(c clip.Clip) string { return p.speech[c.ID()] }

func (p *Page) FindClip(kind clip.Kind, key string) (clip.Clip, bool) {
	want := clip.Clip{Kind: kind, Key: key}.Normalized()
	for i := range p.clips {
		if p.clips[i].Normalized() == want {
			return p.clips[i], true
		}
	}
	return clip.Clip{}, false
}

func (p *Page) collectClips() []clip.Clip {
	var clips []clip.Clip
	seen := map[clip.Clip]bool{}
	add := func(c clip.Clip) {
		if c.Key != "" && !seen[c] {
			seen[c] = true
			clips = append(clips, c)
		}
	}
	for i := range p.Sections {
		blocks := p.Sections[i].Blocks
		for j := range blocks {
			letters := blocks[j].Letters
			for k := range letters {
				add(clip.Clip{Kind: clip.Sound, Key: letters[k].Sound})
				add(clip.Clip{Kind: clip.Word, Key: letters[k].Example})
			}
		}
	}
	p.eachText(func(t Text) {
		found := clipsIn(t.primary())
		for i := range found {
			add(found[i])
		}
	})
	return clips
}

func (p *Page) eachText(fn func(Text)) {
	optional := func(t *Text) {
		if t != nil {
			fn(*t)
		}
	}
	all := func(ts []Text) {
		for i := range ts {
			fn(ts[i])
		}
	}
	fn(p.Title)
	for i := range p.Sections {
		s := &p.Sections[i]
		optional(s.Title)
		for j := range s.Blocks {
			b := &s.Blocks[j]
			optional(b.Text)
			all(b.Items)
			all(b.Columns)
			for k := range b.Rows {
				all(b.Rows[k])
			}
			for k := range b.Letters {
				fn(b.Letters[k].IPA)
				fn(b.Letters[k].Spanish)
				fn(b.Letters[k].English)
			}
		}
	}
}
