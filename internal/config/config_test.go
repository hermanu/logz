package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hermanu/logz/internal/config"
	"github.com/hermanu/logz/internal/parser"
)

func TestDefault(t *testing.T) {
	t.Parallel()
	d := config.Default()
	if d.Output != "pretty" {
		t.Errorf("Output: got %q, want pretty", d.Output)
	}
	if d.Top != 10 {
		t.Errorf("Top: got %d, want 10", d.Top)
	}
	if d.Parser.TextPattern == "" {
		t.Error("TextPattern should have a default")
	}
	if len(d.Parser.LevelKeys) == 0 {
		t.Error("LevelKeys should have defaults")
	}
}

func TestLoadFile_Override(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "logz.yaml")
	body := `
output: json
top: 5
parser:
  level_keys: [severity]
  level_aliases:
    "30": info
    "40": warn
`
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Output != "json" {
		t.Errorf("Output: got %q, want json", cfg.Output)
	}
	if cfg.Top != 5 {
		t.Errorf("Top: got %d, want 5", cfg.Top)
	}
	if len(cfg.Parser.LevelKeys) != 1 || cfg.Parser.LevelKeys[0] != "severity" {
		t.Errorf("LevelKeys: got %v", cfg.Parser.LevelKeys)
	}

	opts, err := cfg.ParserOptions()
	if err != nil {
		t.Fatal(err)
	}
	if opts.LevelAliases["30"] != parser.LevelInfo {
		t.Errorf("alias 30: got %v, want info", opts.LevelAliases["30"])
	}
}

func TestParserOptions_BadAlias(t *testing.T) {
	t.Parallel()
	cfg := config.Default()
	cfg.Parser.LevelAliases = map[string]string{"x": "not-a-level"}
	if _, err := cfg.ParserOptions(); err == nil {
		t.Error("expected error for bad alias")
	}
}

func TestLoad_Env(t *testing.T) {
	t.Setenv("LOGZ_OUTPUT", "json")
	t.Setenv("LOGZ_TOP", "7")
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Output != "json" {
		t.Errorf("Output: got %q, want json (from env)", cfg.Output)
	}
	if cfg.Top != 7 {
		t.Errorf("Top: got %d, want 7", cfg.Top)
	}
}
