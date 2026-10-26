package audioconv

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type Converter struct {
	WAVSampleRate int
	AACBitrate    int
}

func (c Converter) ToWAV(ctx context.Context, in, out string) error {
	return run(ctx, in, out, "-f", "WAVE", "-d", "LEI16@"+strconv.Itoa(c.WAVSampleRate), "-c", "1")
}

func (c Converter) ToM4A(ctx context.Context, in, out string) error {
	return run(ctx, in, out, "-f", "m4af", "-d", "aac", "-c", "1", "-b", strconv.Itoa(c.AACBitrate))
}

func run(ctx context.Context, in, out string, args ...string) error {
	args = append(args, in, out)
	if output, err := exec.CommandContext(ctx, "afconvert", args...).CombinedOutput(); err != nil {
		return fmt.Errorf("afconvert: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}
