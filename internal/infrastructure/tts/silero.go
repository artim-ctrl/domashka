package tts

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/artim-ctrl/domashka/internal/domain/voiceover"
)

type sileroWorker struct {
	command []string
	env     []string

	mu     sync.Mutex
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Scanner
}

type sileroRequest struct {
	Model   string `json:"model"`
	Speaker string `json:"speaker"`
	Text    string `json:"text"`
	Out     string `json:"out"`
}

type sileroResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error"`
}

func (w *sileroWorker) synthesize(v voiceover.Voice, text, outWAV string) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.cmd == nil {
		if err := w.start(); err != nil {
			return err
		}
	}
	req, err := json.Marshal(sileroRequest{Model: v.Model, Speaker: v.Speaker, Text: text, Out: outWAV})
	if err != nil {
		return err
	}
	if _, err := w.stdin.Write(append(req, '\n')); err != nil {
		w.stop()
		return fmt.Errorf("silero worker is not responding: %w", err)
	}
	if !w.stdout.Scan() {
		w.stop()
		return errors.New("silero worker exited, see the server log for details")
	}
	var resp sileroResponse
	if err := json.Unmarshal(w.stdout.Bytes(), &resp); err != nil {
		return fmt.Errorf("silero worker sent an invalid response: %w", err)
	}
	if !resp.OK {
		return fmt.Errorf("silero: %s", resp.Error)
	}
	return nil
}

func (w *sileroWorker) start() error {
	if _, err := exec.LookPath(w.command[0]); err != nil {
		return fmt.Errorf("silero voices need %s: %w", w.command[0], err)
	}
	cmd := exec.Command(w.command[0], w.command[1:]...)
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), w.env...)
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
	w.cmd, w.stdin, w.stdout = cmd, stdin, bufio.NewScanner(stdout)
	return nil
}

func (w *sileroWorker) stop() {
	if w.cmd == nil {
		return
	}
	_ = w.stdin.Close()
	_ = w.cmd.Process.Kill()
	_ = w.cmd.Wait()
	w.cmd = nil
}

func (w *sileroWorker) close() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.stop()
}
