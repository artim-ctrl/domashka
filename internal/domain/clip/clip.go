package clip

import "strings"

type Kind string

const (
	Sound Kind = "s"
	Word  Kind = "w"
)

func ParseKind(s string) (Kind, bool) {
	switch k := Kind(s); k {
	case Sound, Word:
		return k, true
	default:
		return "", false
	}
}

type Dirs struct {
	Sounds string
	Words  string
}

func (d Dirs) For(k Kind) string {
	if k == Sound {
		return d.Sounds
	}
	return d.Words
}

type Clip struct {
	Kind Kind
	Key  string
}

func (c Clip) ID() string { return string(c.Kind) + ":" + c.Key }

func (c Clip) Normalized() Clip { return Clip{Kind: c.Kind, Key: NormalizeKey(c.Key)} }

var keyNormalizer = strings.NewReplacer(
	"й", "й", "Й", "Й",
	"ё", "ё", "Ё", "Ё",
	"́", "",
)

func NormalizeKey(s string) string {
	return strings.ToLower(keyNormalizer.Replace(strings.TrimSpace(s)))
}
