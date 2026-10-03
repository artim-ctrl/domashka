package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestTextCheck(t *testing.T) {
	for _, tc := range []struct {
		json    string
		wantErr string
	}{
		{`"{{w:база}} — {{w:ваза}}"`, ""},
		{`{"es": "{{w:дом}} casa", "en": "{{w:дом}} house", "ru": "{{w:дом}} дом"}`, ""},
		{`{"es": "casa", "en": "house"}`, `нет перевода на "ru"`},
		{`{"es": "{{w:дом}} casa", "en": "house", "ru": "{{w:дом}} дом"}`, `другие кнопки`},
		{`{"es": "a", "en": "a", "ru": "a", "de": "a"}`, `неизвестный язык "de"`},
	} {
		var txt Text
		if err := json.Unmarshal([]byte(tc.json), &txt); err != nil {
			t.Fatal(err)
		}
		err := txt.check()
		switch {
		case tc.wantErr == "" && err != nil:
			t.Errorf("%s: лишняя ошибка %v", tc.json, err)
		case tc.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tc.wantErr)):
			t.Errorf("%s: ждали ошибку %q, получили %v", tc.json, tc.wantErr, err)
		}
	}
}

func TestTextRoundTrip(t *testing.T) {
	for _, in := range []string{`"—"`, `{"en":"b","es":"a","ru":"в"}`} {
		var txt Text
		if err := json.Unmarshal([]byte(in), &txt); err != nil {
			t.Fatal(err)
		}
		out, _ := json.Marshal(txt)
		if string(out) != in {
			t.Errorf("%s → %s", in, out)
		}
	}
}
