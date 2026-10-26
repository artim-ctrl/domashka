package content

import (
	"encoding/json"
	"maps"
	"os"
	"strings"
	"testing"

	"github.com/artim-ctrl/domashka/internal/domain/clip"
)

const (
	repoPage   = "../../../content/page.json"
	repoSpeech = "../../../content/speech.json"
)

func loadRepoContent() (*Page, error) {
	pageData, err := os.ReadFile(repoPage)
	if err != nil {
		return nil, err
	}
	speechData, err := os.ReadFile(repoSpeech)
	if err != nil {
		return nil, err
	}
	return Parse(pageData, speechData)
}

func TestLoad(t *testing.T) {
	p, err := loadRepoContent()
	if err != nil {
		t.Fatal(err)
	}
	ids := map[string]bool{}
	clips := p.Clips()
	for i := range clips {
		ids[clips[i].ID()] = true
	}
	want := []string{"s:ш", "s:нь", "w:мама", "w:молоко", "w:щётка"}
	for i := range want {
		if !ids[want[i]] {
			t.Errorf("page has no clip %s", want[i])
		}
	}
	if ids["w:молоко́"] {
		t.Error("the label after | leaked into the clip key")
	}
}

func TestFindClip(t *testing.T) {
	p, err := loadRepoContent()
	if err != nil {
		t.Fatal(err)
	}
	if c, ok := p.FindClip(clip.Word, "Молоко́"); !ok || c.Key != "молоко" {
		t.Errorf("FindClip ignores case and stress: got %+v, %v", c, ok)
	}
	if _, ok := p.FindClip(clip.Word, "кошка"); ok {
		t.Error("FindClip found a word that is not on the page")
	}
}

func TestValidateSpeech(t *testing.T) {
	p, err := loadRepoContent()
	if err != nil {
		t.Fatal(err)
	}
	cases := map[string]struct {
		edit    func(map[string]string)
		wantErr string
	}{
		"missing word":        {func(m map[string]string) { delete(m, "w:дом") }, "no speech text for w:дом"},
		"typo in word":        {func(m map[string]string) { m["w:дом"] = "д+оом" }, "is not the same word"},
		"lone consonant":      {func(m map[string]string) { m["s:ш"] = "ш" }, "letter name"},
		"extra entry":         {func(m map[string]string) { m["w:кошка"] = "к+ошка" }, "w:кошка is not on the page"},
		"stress mark is fine": {func(m map[string]string) { m["w:дом"] = "д+ом" }, ""},
	}
	for name := range cases {
		tc := cases[name]
		t.Run(name, func(t *testing.T) {
			speech := maps.Clone(p.speech)
			tc.edit(speech)
			err := validateSpeech(p.Clips(), speech)
			if tc.wantErr == "" && err != nil || tc.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tc.wantErr)) {
				t.Errorf("want %q, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestTextValidate(t *testing.T) {
	cases := []struct {
		json    string
		wantErr string
	}{
		{`"{{w:база}} — {{w:ваза}}"`, ""},
		{`{"es": "{{w:дом}} casa", "en": "{{w:дом}} house", "ru": "{{w:дом}} дом"}`, ""},
		{`{"es": "casa", "en": "house"}`, `missing "ru" translation`},
		{`{"es": "{{w:дом}} casa", "en": "house", "ru": "{{w:дом}} дом"}`, `different clips`},
		{`{"es": "a", "en": "a", "ru": "a", "de": "a"}`, `unknown language "de"`},
	}
	for i := range cases {
		tc := cases[i]
		var txt Text
		if err := json.Unmarshal([]byte(tc.json), &txt); err != nil {
			t.Fatal(err)
		}
		err := txt.validate()
		switch {
		case tc.wantErr == "" && err != nil:
			t.Errorf("%s: unexpected error %v", tc.json, err)
		case tc.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tc.wantErr)):
			t.Errorf("%s: want error %q, got %v", tc.json, tc.wantErr, err)
		}
	}
}

func TestTextRoundTrip(t *testing.T) {
	inputs := []string{`"—"`, `{"en":"b","es":"a","ru":"в"}`}
	for i := range inputs {
		var txt Text
		if err := json.Unmarshal([]byte(inputs[i]), &txt); err != nil {
			t.Fatal(err)
		}
		out, err := json.Marshal(txt)
		if err != nil {
			t.Fatal(err)
		}
		if string(out) != inputs[i] {
			t.Errorf("%s → %s", inputs[i], out)
		}
	}
}

func TestUnknownBlockType(t *testing.T) {
	_, err := Parse([]byte(`{"title":"x","sections":[{"id":"a","blocks":[{"type":"video"}]}]}`), []byte(`{}`))
	if err == nil || !strings.Contains(err.Error(), `unknown block type "video"`) {
		t.Errorf("want unknown block type error, got %v", err)
	}
}
