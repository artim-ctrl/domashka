package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPageLoads(t *testing.T) {
	p, err := loadPage()
	if err != nil {
		t.Fatal(err)
	}
	ids := map[string]bool{}
	for _, c := range p.Clips() {
		ids[c.ID()] = true
	}
	for _, want := range []string{"s:ш", "s:нь", "w:мама", "w:молоко", "w:щётка"} {
		if !ids[want] {
			t.Errorf("на странице нет записи %s", want)
		}
	}
	if ids["w:молоко́"] {
		t.Error("подпись после | попала в ключ записи")
	}
}

func TestScanNormalizesFileNames(t *testing.T) {
	dir := t.TempDir()
	words := filepath.Join(dir, "words")
	if err := os.MkdirAll(words, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{
		"ёлка.mp3",           // как есть
		"щётка.m4a",         // ё разложена: е + U+0308
		"май.wav", "мӑ.m4a", // мусор рядом не мешает
		"Дом.mp3", "дом.m4a", // .m4a предпочтительнее
		"молоко́.ogg", // ударение в имени
		"заметки.txt",
	} {
		if err := os.WriteFile(filepath.Join(words, name), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	lib := &Library{dir: dir}
	for key, wantFile := range map[string]string{
		"ёлка":   "ёлка.mp3",
		"щётка":  "щётка.m4a",
		"май":    "май.wav",
		"дом":    "дом.m4a",
		"молоко": "молоко́.ogg",
	} {
		rec, ok, err := lib.Lookup(Clip{Kind: KindWord, Key: key})
		if err != nil || !ok {
			t.Errorf("%s: не нашлось (err=%v)", key, err)
			continue
		}
		if got := filepath.Base(rec.Path); got != wantFile {
			t.Errorf("%s: взят %q, ждали %q", key, got, wantFile)
		}
	}
	if _, ok, _ := lib.Lookup(Clip{Kind: KindWord, Key: "заметки"}); ok {
		t.Error("txt принят за запись")
	}
}
