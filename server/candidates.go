package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// Варианты озвучки лежат в candidates/<sounds|words>/<ключ>/<голос>.wav (не в git),
// рядом text.txt — каким текстом их делали. Выбранный вариант становится записью в audio/.

type Candidates struct {
	dir string
}

type Candidate struct {
	Voice string `json:"voice"`
	URL   string `json:"url"`
}

func (cs *Candidates) clipDir(c Clip) string { return filepath.Join(cs.dir, c.Kind.dir(), c.Key) }

func (cs *Candidates) Path(c Clip, voice string) string {
	return filepath.Join(cs.clipDir(c), voice+".wav")
}

// List — готовые варианты в порядке голосов и текст, которым их делали (пусто — ещё не делали).
func (cs *Candidates) List(c Clip) ([]Candidate, string, error) {
	out := []Candidate{}
	for _, v := range voices {
		info, err := os.Stat(cs.Path(c, v.ID))
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, "", err
		}
		out = append(out, Candidate{
			Voice: v.ID,
			URL:   fmt.Sprintf("/api/admin/candidates/%s/%s/%s?v=%d", c.Kind, url.PathEscape(c.Key), v.ID, info.ModTime().UnixMilli()),
		})
	}
	text, err := os.ReadFile(filepath.Join(cs.clipDir(c), "text.txt"))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, "", err
	}
	return out, string(text), nil
}

func (cs *Candidates) Generate(tts *TTS, c Clip, v Voice, text string) (Candidate, error) {
	dir := cs.clipDir(c)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return Candidate{}, err
	}
	if err := os.WriteFile(filepath.Join(dir, "text.txt"), []byte(text), 0o644); err != nil {
		return Candidate{}, err
	}
	if err := tts.Synthesize(v, text, cs.Path(c, v.ID)); err != nil {
		return Candidate{}, err
	}
	list, _, err := cs.List(c)
	if err != nil {
		return Candidate{}, err
	}
	for _, cand := range list {
		if cand.Voice == v.ID {
			return cand, nil
		}
	}
	return Candidate{}, errors.New("вариант не появился на диске")
}

// Choice — каким голосом и текстом сделана запись в audio/. Хранится в audio/choices.json.
type Choice struct {
	Voice string `json:"voice"`
	Text  string `json:"text"`
}

func (l *Library) choicesPath() string { return filepath.Join(l.dir, "choices.json") }

func (l *Library) Choices() (map[string]Choice, error) {
	choices := map[string]Choice{}
	b, err := os.ReadFile(l.choicesPath())
	if errors.Is(err, os.ErrNotExist) {
		return choices, nil
	}
	if err != nil {
		return nil, err
	}
	return choices, json.Unmarshal(b, &choices)
}

// Choose делает вариант записью: AAC в audio/<…>/<ключ>.m4a (маленький, играет везде,
// включая Safari на iPhone) и отметка в choices.json.
func (l *Library) Choose(c Clip, wav string, choice Choice) (Recording, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	dir := filepath.Join(l.dir, c.Kind.dir())
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return Recording{}, err
	}
	part := filepath.Join(dir, ".choose-"+c.Key+".tmp")
	if err := afconvert(wav, part, "-f", "m4af", "-d", "aac", "-c", "1", "-b", "64000"); err != nil {
		os.Remove(part)
		return Recording{}, err
	}
	final := filepath.Join(dir, c.Key+".m4a")
	if err := os.Rename(part, final); err != nil {
		return Recording{}, err
	}
	if err := removeOtherFormats(dir, c.Key, final); err != nil {
		return Recording{}, err
	}

	choices, err := l.Choices()
	if err != nil {
		return Recording{}, err
	}
	choices[c.ID()] = choice
	b, _ := json.MarshalIndent(choices, "", "  ")
	if err := os.WriteFile(l.choicesPath(), append(b, '\n'), 0o644); err != nil {
		return Recording{}, err
	}

	info, err := os.Stat(final)
	if err != nil {
		return Recording{}, err
	}
	return Recording{Path: final, Version: info.ModTime().UnixMilli()}, nil
}

// removeOtherFormats убирает прежние файлы записи в других форматах или с другим
// написанием имени — иначе, например, старый «Дом.wav» спорил бы с новым «дом.m4a».
func removeOtherFormats(dir, key, keep string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	var errs []error
	for _, e := range entries {
		name := e.Name()
		path := filepath.Join(dir, name)
		if path == keep || e.IsDir() || extRank(strings.ToLower(filepath.Ext(name))) < 0 {
			continue
		}
		if normalizeKey(strings.TrimSuffix(name, filepath.Ext(name))) == normalizeKey(key) {
			errs = append(errs, os.Remove(path))
		}
	}
	return errors.Join(errs...)
}
