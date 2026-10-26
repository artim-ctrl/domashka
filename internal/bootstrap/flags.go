package bootstrap

import (
	"errors"
	"flag"
	"io"
)

const defaultConfig = "config/base.yaml"

type mode int

const (
	modeServe mode = iota
	modeHealthcheck
	modeMissing
	modeVoices
	modeVoicesAll
)

type flags struct {
	configs []string
	admin   bool
	mode    mode
}

func parseFlags(args []string, output io.Writer) (flags, error) {
	var (
		f                                       flags
		healthcheck, missing, voices, voicesAll bool
	)
	fs := flag.NewFlagSet("domashka", flag.ContinueOnError)
	fs.SetOutput(output)
	fs.Func("config", "YAML config file; repeat to layer files, later ones override earlier ones (default "+defaultConfig+")", func(path string) error {
		if path == "" {
			return errors.New("empty config path")
		}
		f.configs = append(f.configs, path)
		return nil
	})
	fs.BoolVar(&f.admin, "admin", false, "enable the admin mode: generating and choosing voiceovers (macOS only)")
	fs.BoolVar(&missing, "missing", false, "print which recordings are missing and exit")
	fs.BoolVar(&voices, "voices", false, "generate candidates with every voice where they are missing and exit")
	fs.BoolVar(&voicesAll, "voices-all", false, "regenerate all candidates and exit")
	fs.BoolVar(&healthcheck, "healthcheck", false, "check that the configured server responds and exit (for Docker)")
	if err := fs.Parse(args); err != nil {
		return flags{}, err
	}
	if fs.NArg() > 0 {
		return flags{}, errors.New("unexpected arguments: " + fs.Arg(0))
	}
	if len(f.configs) == 0 {
		f.configs = []string{defaultConfig}
	}

	switch {
	case healthcheck:
		f.mode = modeHealthcheck
	case missing:
		f.mode = modeMissing
	case voicesAll:
		f.mode = modeVoicesAll
	case voices:
		f.mode = modeVoices
	default:
		f.mode = modeServe
	}
	return f, nil
}
