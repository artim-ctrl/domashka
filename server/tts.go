package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
)

// Voice — голос, которым можно озвучить запись.
type Voice struct {
	ID      string `json:"id"`
	Label   string `json:"label"`
	License string `json:"license"`

	engine  string // silero | say
	model   string // модель Silero или голос say
	speaker string
}

// Голоса для сравнения. Silero v5_5_ru — носители языка, но лицензия некоммерческая;
// v5_cis_base — MIT, только дикторы там из других языков СНГ, возможен акцент.
var voices = []Voice{
	{ID: "silero-xenia", Label: "Silero · Xenia", License: "CC BY-NC", engine: "silero", model: "v5_5_ru", speaker: "xenia"},
	{ID: "silero-baya", Label: "Silero · Baya", License: "CC BY-NC", engine: "silero", model: "v5_5_ru", speaker: "baya"},
	{ID: "silero-kseniya", Label: "Silero · Kseniya", License: "CC BY-NC", engine: "silero", model: "v5_5_ru", speaker: "kseniya"},
	{ID: "silero-aidar", Label: "Silero · Aidar", License: "CC BY-NC", engine: "silero", model: "v5_5_ru", speaker: "aidar"},
	{ID: "silero-eugene", Label: "Silero · Eugene", License: "CC BY-NC", engine: "silero", model: "v5_5_ru", speaker: "eugene"},
	{ID: "mit-ekaterina", Label: "Silero CIS · Ekaterina", License: "MIT", engine: "silero", model: "v5_cis_base", speaker: "ru_ekaterina"},
	{ID: "mit-zinaida", Label: "Silero CIS · Zinaida", License: "MIT", engine: "silero", model: "v5_cis_base", speaker: "ru_zinaida"},
	{ID: "mit-dmitriy", Label: "Silero CIS · Dmitriy", License: "MIT", engine: "silero", model: "v5_cis_base", speaker: "ru_dmitriy"},
	{ID: "mit-igor", Label: "Silero CIS · Igor", License: "MIT", engine: "silero", model: "v5_cis_base", speaker: "ru_igor"},
	{ID: "macos-milena", Label: "macOS · Milena", License: "macOS", engine: "say", model: "Milena"},
}

func voiceByID(id string) (Voice, bool) {
	for _, v := range voices {
		if v.ID == id {
			return v, true
		}
	}
	return Voice{}, false
}

// TTS озвучивает текст в WAV. Silero живёт в отдельном процессе Python (tools/tts.py):
// он запускается при первой генерации и держит модели в памяти.
type TTS struct {
	script string

	mu     sync.Mutex
	worker *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Scanner
}

func (t *TTS) Synthesize(v Voice, text, outWav string) error {
	if v.engine == "say" {
		return say(v.model, strings.ReplaceAll(text, "+", ""), outWav)
	}
	return t.silero(v, text, outWav)
}

func (t *TTS) silero(v Voice, text, outWav string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.worker == nil {
		if err := t.start(); err != nil {
			return err
		}
	}
	req, _ := json.Marshal(map[string]string{"model": v.model, "speaker": v.speaker, "text": text, "out": outWav})
	if _, err := t.stdin.Write(append(req, '\n')); err != nil {
		t.stop()
		return fmt.Errorf("синтезатор Silero не отвечает: %w", err)
	}
	if !t.stdout.Scan() {
		t.stop()
		return errors.New("синтезатор Silero упал — подробности в логе сервера")
	}
	var resp struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
	}
	if err := json.Unmarshal(t.stdout.Bytes(), &resp); err != nil || !resp.OK {
		return fmt.Errorf("Silero: %s", resp.Error)
	}
	return nil
}

func (t *TTS) start() error {
	if _, err := exec.LookPath("uv"); err != nil {
		return errors.New("для голосов Silero нужен uv (https://docs.astral.sh/uv/)")
	}
	cmd := exec.Command("uv", "run", "--quiet", "--script", t.script)
	cmd.Stderr = os.Stderr // загрузка моделей и ошибки Python — в лог сервера
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	t.worker, t.stdin, t.stdout = cmd, stdin, bufio.NewScanner(stdout)
	return nil
}

func (t *TTS) stop() {
	if t.worker != nil {
		t.stdin.Close()
		_ = t.worker.Process.Kill()
		_ = t.worker.Wait()
		t.worker = nil
	}
}

func (t *TTS) Close() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.stop()
}

func say(voice, text, outWav string) error {
	aiff := outWav + ".aiff"
	defer os.Remove(aiff)
	if out, err := exec.Command("say", "-v", voice, "-r", "150", "-o", aiff, "--", text).CombinedOutput(); err != nil {
		return fmt.Errorf("say: %v: %s", err, strings.TrimSpace(string(out)))
	}
	return afconvert(aiff, outWav, "-f", "WAVE", "-d", "LEI16@24000", "-c", "1")
}

func afconvert(in, out string, args ...string) error {
	if o, err := exec.Command("afconvert", append(append(args, in), out)...).CombinedOutput(); err != nil {
		return fmt.Errorf("afconvert: %v: %s", err, strings.TrimSpace(string(o)))
	}
	return nil
}
