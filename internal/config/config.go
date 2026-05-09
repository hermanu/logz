// Package config loads logz configuration from files, environment, and code
// defaults. Higher-priority sources override lower-priority ones in this
// order: built-in defaults < user config < project config < env vars < flags.
//
// Flag overrides are applied by the cmd layer after [Load] returns.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"

	"github.com/hermanu/logz/internal/parser"
)

// Config is the user-tunable configuration surface.
type Config struct {
	// Format pins a parser format ("json", "text"). Empty means auto-detect.
	Format string `koanf:"format"`

	// Output mode: "pretty" or "json".
	Output string `koanf:"output"`

	// NoColor disables ANSI colors. Equivalent to --no-color.
	NoColor bool `koanf:"no_color"`

	// Top is the default top-N for `logz summary`.
	Top int `koanf:"top"`

	// Parser holds parser-specific knobs.
	Parser ParserConfig `koanf:"parser"`

	// Theme is the optional named color theme. Reserved for future use.
	Theme string `koanf:"theme"`
}

// ParserConfig mirrors [parser.Options] in a YAML-friendly shape. Each field
// is optional; an empty slice or zero value falls back to the corresponding
// built-in default.
type ParserConfig struct {
	// LevelKeys are the structured-field keys to inspect for log level, in
	// priority order. The first key present in the entry wins.
	LevelKeys []string `koanf:"level_keys"`

	// MessageKeys are the structured-field keys to inspect for the human
	// message, in priority order.
	MessageKeys []string `koanf:"message_keys"`

	// TimestampKeys are the structured-field keys to inspect for timestamps,
	// in priority order.
	TimestampKeys []string `koanf:"timestamp_keys"`

	// TimestampLayouts are extra [time.Parse] layouts to attempt before falling
	// back to the package's built-in list (RFC3339 et al.).
	TimestampLayouts []string `koanf:"timestamp_layouts"`

	// LevelAliases maps custom level strings (lowercased) to canonical level
	// names. Useful for Bunyan-style numeric levels (e.g. "30" → "info").
	LevelAliases map[string]string `koanf:"level_aliases"`

	// TextPattern is a Go regexp with named capture groups used by the plain
	// text parser. Recognized group names are "timestamp", "level", and
	// "message"; any other named group becomes a structured field.
	TextPattern string `koanf:"text_pattern"`
}

// Default returns a Config populated with built-in defaults.
func Default() Config {
	d := parser.DefaultOptions()
	return Config{
		Output: "pretty",
		Top:    10,
		Parser: ParserConfig{
			LevelKeys:     d.LevelKeys,
			MessageKeys:   d.MessageKeys,
			TimestampKeys: d.TimestampKeys,
			TextPattern:   parser.DefaultTextPattern,
		},
	}
}

// Load reads configuration from the standard locations.
//
// Search order (later wins):
//  1. Built-in defaults
//  2. $XDG_CONFIG_HOME/logz/config.yaml (or ~/.config/logz/config.yaml)
//  3. ./.logz.yaml in the current directory
//  4. Environment variables prefixed LOGZ_ (e.g. LOGZ_OUTPUT=json,
//     LOGZ_PARSER_LEVEL_KEYS=level,severity)
//
// Missing files are not an error — only malformed ones.
func Load() (Config, error) {
	k := koanf.New(".")
	cfg := Default()

	// Defaults via struct round-trip.
	if err := k.Load(structProvider(cfg), nil); err != nil {
		return cfg, fmt.Errorf("config: load defaults: %w", err)
	}

	for _, path := range searchPaths() {
		if _, err := os.Stat(path); err != nil {
			continue
		}
		if err := k.Load(file.Provider(path), yaml.Parser()); err != nil {
			return cfg, fmt.Errorf("config: load %s: %w", path, err)
		}
	}

	// Env vars: LOGZ_PARSER_LEVEL_KEYS=level,severity → parser.level_keys.
	envProvider := env.Provider("LOGZ_", ".", func(s string) string {
		s = strings.TrimPrefix(s, "LOGZ_")
		s = strings.ToLower(s)
		return strings.ReplaceAll(s, "_", ".")
	})
	if err := k.Load(envProvider, nil); err != nil {
		return cfg, fmt.Errorf("config: load env: %w", err)
	}

	if err := k.Unmarshal("", &cfg); err != nil {
		return cfg, fmt.Errorf("config: unmarshal: %w", err)
	}
	return cfg, nil
}

// LoadFile reads configuration from a single explicit file plus environment.
// Used when the user passes --config to override the default search.
func LoadFile(path string) (Config, error) {
	k := koanf.New(".")
	cfg := Default()

	if err := k.Load(structProvider(cfg), nil); err != nil {
		return cfg, fmt.Errorf("config: load defaults: %w", err)
	}
	if err := k.Load(file.Provider(path), yaml.Parser()); err != nil {
		return cfg, fmt.Errorf("config: load %s: %w", path, err)
	}
	envProvider := env.Provider("LOGZ_", ".", func(s string) string {
		s = strings.TrimPrefix(s, "LOGZ_")
		s = strings.ToLower(s)
		return strings.ReplaceAll(s, "_", ".")
	})
	if err := k.Load(envProvider, nil); err != nil {
		return cfg, fmt.Errorf("config: load env: %w", err)
	}
	if err := k.Unmarshal("", &cfg); err != nil {
		return cfg, fmt.Errorf("config: unmarshal: %w", err)
	}
	return cfg, nil
}

// ParserOptions converts the YAML-friendly ParserConfig into runtime
// [parser.Options]. Invalid level aliases are reported.
func (c Config) ParserOptions() (parser.Options, error) {
	opts := parser.Options{
		LevelKeys:        c.Parser.LevelKeys,
		MessageKeys:      c.Parser.MessageKeys,
		TimestampKeys:    c.Parser.TimestampKeys,
		TimestampLayouts: c.Parser.TimestampLayouts,
	}
	if len(c.Parser.LevelAliases) > 0 {
		opts.LevelAliases = make(map[string]parser.Level, len(c.Parser.LevelAliases))
		for k, v := range c.Parser.LevelAliases {
			lvl, ok := parser.ParseLevel(v)
			if !ok {
				return opts, fmt.Errorf("config: level alias %q → %q: unknown target level", k, v)
			}
			opts.LevelAliases[strings.ToLower(k)] = lvl
		}
	}
	return opts, nil
}

func searchPaths() []string {
	var paths []string

	xdg := os.Getenv("XDG_CONFIG_HOME")
	if xdg == "" {
		if home, err := os.UserHomeDir(); err == nil {
			xdg = filepath.Join(home, ".config")
		}
	}
	if xdg != "" {
		paths = append(paths, filepath.Join(xdg, "logz", "config.yaml"))
	}
	paths = append(paths, ".logz.yaml")
	return paths
}

// structProvider is a tiny koanf provider that seeds defaults from a struct.
// Keeping this inline avoids pulling in koanf/providers/structs.
type structProvider Config

func (s structProvider) ReadBytes() ([]byte, error) {
	return nil, errors.New("structProvider does not provide raw bytes")
}

func (s structProvider) Read() (map[string]any, error) {
	return map[string]any{
		"format":   s.Format,
		"output":   s.Output,
		"no_color": s.NoColor,
		"top":      s.Top,
		"theme":    s.Theme,
		"parser": map[string]any{
			"level_keys":        s.Parser.LevelKeys,
			"message_keys":      s.Parser.MessageKeys,
			"timestamp_keys":    s.Parser.TimestampKeys,
			"timestamp_layouts": s.Parser.TimestampLayouts,
			"level_aliases":     s.Parser.LevelAliases,
			"text_pattern":      s.Parser.TextPattern,
		},
	}, nil
}
