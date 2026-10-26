package clip

import "testing"

func TestNormalizeKey(t *testing.T) {
	cases := map[string]string{
		" Дом ":   "дом",
		"щётка":  "щётка",
		"май":    "май",
		"молоко́": "молоко",
	}
	for in := range cases {
		if got := NormalizeKey(in); got != cases[in] {
			t.Errorf("NormalizeKey(%q) = %q, want %q", in, got, cases[in])
		}
	}
}

func TestParseKind(t *testing.T) {
	cases := map[string]bool{"s": true, "w": true, "x": false, "": false}
	for in := range cases {
		if _, ok := ParseKind(in); ok != cases[in] {
			t.Errorf("ParseKind(%q) ok = %v, want %v", in, ok, cases[in])
		}
	}
}
