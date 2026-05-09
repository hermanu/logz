// Package cmd wires the CLI surface for logz.
package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/hermanu/logz/internal/config"
	"github.com/hermanu/logz/internal/logz"
)

// ExitCode follows grep's convention: 0 = match, 1 = no match, 2 = error.
type ExitCode int

// Documented exit codes.
const (
	ExitMatched   ExitCode = 0
	ExitNoMatches ExitCode = 1
	ExitError     ExitCode = 2
)

// errNoMatches is returned by commands when their filter matched nothing.
// It's not a "real" error but lets us route through the standard error path
// to set the right exit code.
var errNoMatches = errors.New("no matching log lines")

// globalFlags are bound on the root command; subcommands read them via [getGlobals].
type globalFlags struct {
	format  string
	output  string
	noColor bool
	cfgPath string
}

// New builds the root command. stdout/stderr/stdin are injected for
// testability — the entrypoint passes the real os.* values.
func New(stdout, stderr io.Writer, stdin io.Reader) *cobra.Command {
	g := &globalFlags{}

	root := &cobra.Command{
		Use:           "logz",
		Short:         "Parse, filter, and summarize log files",
		Long:          "logz is a fast CLI for parsing, filtering, and summarizing log files in JSON, logfmt, plain text, and Apache/Nginx formats.",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       fmt.Sprintf("%s (commit %s, built %s)", logz.Version(), logz.Commit(), logz.Date()),
	}

	root.PersistentFlags().StringVar(&g.format, "format", "", "force log format: json, text (default: auto-detect)")
	root.PersistentFlags().StringVarP(&g.output, "output", "o", "", "output mode: pretty, json (default: pretty)")
	root.PersistentFlags().BoolVar(&g.noColor, "no-color", false, "disable colored output")
	root.PersistentFlags().StringVar(&g.cfgPath, "config", "", "path to a config file (overrides default search paths)")

	root.SetOut(stdout)
	root.SetErr(stderr)
	root.SetIn(stdin)

	// Stash globals on the root so subcommands can read them.
	root.SetContext(withGlobals(root.Context(), g))

	root.AddCommand(
		newFilterCmd(),
		newSummaryCmd(),
		newTailCmd(),
		newFieldsCmd(),
	)
	return root
}

// Run is the binary entrypoint. It returns the appropriate ExitCode and prints
// errors to stderr.
func Run(args []string, stdout, stderr io.Writer, stdin io.Reader) ExitCode {
	root := New(stdout, stderr, stdin)
	root.SetArgs(args)

	err := root.Execute()
	switch {
	case err == nil:
		return ExitMatched
	case errors.Is(err, errNoMatches):
		return ExitNoMatches
	default:
		fmt.Fprintf(stderr, "logz: %v\n", err)
		return ExitError
	}
}

// loadConfig honors --config when set, otherwise falls back to the standard
// search paths used by [config.Load].
func loadConfig(path string) (config.Config, error) {
	if path == "" {
		return config.Load()
	}
	if _, err := os.Stat(path); err != nil {
		return config.Config{}, fmt.Errorf("config: %w", err)
	}
	// Setting LOGZ_CONFIG_FILE before Load() is overkill; callers requesting an
	// explicit path get a focused load that reads only that file plus env.
	return config.LoadFile(path)
}
