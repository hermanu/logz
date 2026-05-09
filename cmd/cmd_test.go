package cmd_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hermanu/logz/cmd"
)

func runCLI(t *testing.T, stdin string, args ...string) (string, cmd.ExitCode) {
	t.Helper()
	var out bytes.Buffer
	code := cmd.Run(args, &out, &out, strings.NewReader(stdin))
	return out.String(), code
}

func TestFilter_JSONFromStdin(t *testing.T) {
	t.Parallel()
	input := `{"level":"info","msg":"hello","ts":"2024-01-01T12:00:00Z"}
`
	stdout, code := runCLI(t, input, "filter")
	if code != 0 {
		t.Fatalf("filter failed: %s", stdout)
	}
	if !strings.Contains(stdout, "hello") {
		t.Errorf("expected 'hello', got %q", stdout)
	}
}

func TestFilter_InvalidJSON(t *testing.T) {
	t.Parallel()
	input := `not json
`
	stdout, code := runCLI(t, input, "filter")
	if code == 0 {
		t.Error("expected failure for invalid JSON")
	}
	if !strings.Contains(stdout, "parse") {
		t.Errorf("expected parse error, got %q", stdout)
	}
}