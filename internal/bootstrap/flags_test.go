package bootstrap

import (
	"io"
	"testing"
)

func TestParseFlagsMode(t *testing.T) {
	cases := map[string]struct {
		args []string
		want mode
	}{
		"serve by default": {nil, modeServe},
		"healthcheck":      {[]string{"-healthcheck"}, modeHealthcheck},
		"missing":          {[]string{"-missing"}, modeMissing},
		"voices":           {[]string{"-voices"}, modeVoices},
		"voices-all wins":  {[]string{"-voices", "-voices-all"}, modeVoicesAll},
	}
	for name := range cases {
		tc := cases[name]
		t.Run(name, func(t *testing.T) {
			f, err := parseFlags(tc.args, io.Discard)
			if err != nil {
				t.Fatal(err)
			}
			if f.mode != tc.want {
				t.Errorf("mode %d, want %d", f.mode, tc.want)
			}
		})
	}
}

func TestParseFlagsRejectsUnknownFlag(t *testing.T) {
	if _, err := parseFlags([]string{"-nope"}, io.Discard); err == nil {
		t.Error("unknown flag accepted")
	}
}

func TestParseFlagsConfigs(t *testing.T) {
	f, err := parseFlags(nil, io.Discard)
	if err != nil || len(f.configs) != 1 || f.configs[0] != defaultConfig {
		t.Errorf("default configs: %v %v", f.configs, err)
	}
	f, err = parseFlags([]string{"-config", "a.yaml", "-config", "b.yaml"}, io.Discard)
	if err != nil || len(f.configs) != 2 || f.configs[1] != "b.yaml" {
		t.Errorf("layered configs: %v %v", f.configs, err)
	}
}
