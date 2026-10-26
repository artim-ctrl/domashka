package recordings

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/artim-ctrl/domashka/internal/domain/clip"
	"github.com/artim-ctrl/domashka/internal/domain/voiceover"
)

func testOptions(dir string) Options {
	return Options{
		Dir:          dir,
		KindDirs:     clip.Dirs{Sounds: "sounds", Words: "words"},
		ChoicesFile:  "choices.json",
		RecordingExt: ".m4a",
		Formats: []Format{
			{Ext: ".m4a", ContentType: "audio/mp4"},
			{Ext: ".mp3", ContentType: "audio/mpeg"},
			{Ext: ".ogg", ContentType: "audio/ogg"},
			{Ext: ".wav", ContentType: "audio/wav"},
		},
	}
}

func TestChooseNeedsEncoder(t *testing.T) {
	lib := New(testOptions(t.TempDir()))
	if _, err := lib.Choose(t.Context(), clip.Clip{Kind: clip.Word, Key: "дом"}, "in.wav", voiceover.Choice{}); err == nil {
		t.Error("Choose without an encoder succeeded")
	}
}

func TestScanNormalizesFileNames(t *testing.T) {
	dir := t.TempDir()
	words := filepath.Join(dir, "words")
	if err := os.MkdirAll(words, 0o755); err != nil {
		t.Fatal(err)
	}
	files := []string{
		"ёлка.mp3",
		"щётка.m4a",
		"май.wav",
		"мӑ.m4a",
		"Дом.mp3", "дом.m4a",
		"молоко́.ogg",
		"заметки.txt",
	}
	for i := range files {
		if err := os.WriteFile(filepath.Join(words, files[i]), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	lib := New(testOptions(dir))
	want := map[string]string{
		"ёлка":   "ёлка.mp3",
		"щётка":  "щётка.m4a",
		"май":    "май.wav",
		"дом":    "дом.m4a",
		"молоко": "молоко́.ogg",
	}
	for key := range want {
		rec, ok, err := lib.Lookup(clip.Clip{Kind: clip.Word, Key: key})
		if err != nil || !ok {
			t.Errorf("%s: not found (err=%v)", key, err)
			continue
		}
		if got := filepath.Base(rec.Path); got != want[key] {
			t.Errorf("%s: picked %q, want %q", key, got, want[key])
		}
	}
	if _, ok, _ := lib.Lookup(clip.Clip{Kind: clip.Word, Key: "заметки"}); ok {
		t.Error("a .txt file was taken for a recording")
	}
}

func TestFrozenLibrary(t *testing.T) {
	dir := t.TempDir()
	words := filepath.Join(dir, "words")
	if err := os.MkdirAll(words, 0o755); err != nil {
		t.Fatal(err)
	}
	lib := New(testOptions(dir))
	if err := lib.Freeze(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(words, "дом.m4a"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, ok, _ := lib.Lookup(clip.Clip{Kind: clip.Word, Key: "дом"}); ok {
		t.Error("the directory is rescanned after Freeze")
	}
}

func TestContentType(t *testing.T) {
	lib := New(testOptions(t.TempDir()))
	cases := map[string]string{"a.m4a": "audio/mp4", "b.MP3": "audio/mpeg", "c.txt": "application/octet-stream"}
	for name := range cases {
		if got := lib.ContentType(name); got != cases[name] {
			t.Errorf("ContentType(%q) = %q, want %q", name, got, cases[name])
		}
	}
}
