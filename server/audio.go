package main

import (
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Kind string

const (
	KindSound Kind = "s" // audio/sounds/ш.m4a
	KindWord  Kind = "w" // audio/words/дом.m4a
)

func (k Kind) dir() string {
	if k == KindSound {
		return "sounds"
	}
	return "words"
}

type Clip struct {
	Kind Kind
	Key  string
}

func (c Clip) ID() string { return string(c.Kind) + ":" + c.Key }

// Расширения в порядке предпочтения, если одна запись лежит в нескольких форматах.
var audioExts = []string{".m4a", ".mp3", ".ogg", ".opus", ".webm", ".wav", ".aac"}

// Системная таблица macOS отдаёт .m4a как audio/mp4a-latm, а Chrome такое не играет.
func init() {
	for ext, typ := range map[string]string{
		".m4a": "audio/mp4", ".mp3": "audio/mpeg", ".ogg": "audio/ogg", ".opus": "audio/ogg",
		".webm": "audio/webm", ".wav": "audio/wav", ".aac": "audio/aac",
	} {
		_ = mime.AddExtensionType(ext, typ)
	}
}

// Library — папка с записями. Каждый раз читается заново: новая запись видна
// после обновления страницы, без перезапуска сервера.
type Library struct {
	dir string
	mu  sync.Mutex // выбор записи — по одному
}

type Recording struct {
	Path    string
	Version int64 // время изменения файла (мс) — чтобы перезаписанный звук не брался из кэша
}

func (l *Library) Scan() (map[Clip]Recording, error) {
	found := map[Clip]Recording{}
	for _, kind := range []Kind{KindSound, KindWord} {
		dir := filepath.Join(l.dir, kind.dir())
		entries, err := os.ReadDir(dir)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		rank := map[Clip]int{}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			ext := strings.ToLower(filepath.Ext(e.Name()))
			r := extRank(ext)
			if r < 0 {
				continue
			}
			clip := Clip{Kind: kind, Key: normalizeKey(strings.TrimSuffix(e.Name(), filepath.Ext(e.Name())))}
			if prev, ok := rank[clip]; ok && prev <= r {
				continue
			}
			info, err := e.Info()
			if err != nil {
				return nil, err
			}
			rank[clip] = r
			found[clip] = Recording{Path: filepath.Join(dir, e.Name()), Version: info.ModTime().UnixMilli()}
		}
	}
	return found, nil
}

func (l *Library) Lookup(c Clip) (Recording, bool, error) {
	all, err := l.Scan()
	if err != nil {
		return Recording{}, false, err
	}
	rec, ok := all[Clip{Kind: c.Kind, Key: normalizeKey(c.Key)}]
	return rec, ok, nil
}

// ExpectedPath — где появится запись, которой пока нет.
func (l *Library) ExpectedPath(c Clip) string {
	return filepath.Join(l.dir, c.Kind.dir(), c.Key+".m4a")
}

func extRank(ext string) int {
	for i, e := range audioExts {
		if e == ext {
			return i
		}
	}
	return -1
}

// Finder и некоторые программы записи сохраняют «й» и «ё» как букву плюс
// диакритику; ударение (молоко́) в имени файла тоже допустимо.
var keyNormalizer = strings.NewReplacer(
	"й", "й", "Й", "Й",
	"ё", "ё", "Ё", "Ё",
	"́", "",
)

func normalizeKey(s string) string {
	return strings.ToLower(keyNormalizer.Replace(strings.TrimSpace(s)))
}

func printMissing(p *Page, lib *Library) error {
	have, err := lib.Scan()
	if err != nil {
		return err
	}
	clips := p.Clips()
	var missing []Clip
	for _, c := range clips {
		if _, ok := have[Clip{Kind: c.Kind, Key: normalizeKey(c.Key)}]; !ok {
			missing = append(missing, c)
		}
	}
	fmt.Printf("Записано %d из %d.\n", len(clips)-len(missing), len(clips))
	for _, c := range missing {
		fmt.Println("  " + lib.ExpectedPath(c))
	}
	return nil
}
