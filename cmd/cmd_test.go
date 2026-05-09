package cmd_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hermanu/logz/cmd"
)

func runCLI(t *testing.T, stdin string, args ...string) (stdout, stderr string, code cmd.ExitCode) {
	t.Helper()
	var out, err bytes.Buffer
	code = cmd.Run(args, &out, &err, strings.NewReader(stdin))
	return out.String(), err.String(), code
}

func TestFilter_JSONFromStdin(t *testing.T) {
	t.Parallel()
	input := `{"level":"info","msg":"hello","ts":"2024-01-01T12:00:00Z"}
{"level":"error","msg":"boom","ts":"2024-01-01T12:00:01Z"}
`
	out, _, code := runCLI(t, input, "filter", "--level", "error", "--no-color")
	if code != cmd.ExitMatched {
		t.Fatalf("exit code: got %d, want 0\nstdout: %s", code, out)
	}
	if !strings.Contains(out, "boom") {
		t.Errorf("expected match for 'boom':\n%s", out)
	}
	if strings.Contains(out, "hello") {
		t.Errorf("info line should be filtered out:\n%s", out)
	}
}

func TestFilter_NoMatchExitCode(t *testing.T) {
	t.Parallel()
	input := `{"level":"info","msg":"hi"}` + "\n"
	_, _, code := runCLI(t, input, "filter", "--level", "fatal", "--no-color")
	if code != cmd.ExitNoMatches {
		t.Errorf("exit code: got %d, want 1", code)
	}
}

func TestSummary_Smoke(t *testing.T) {
	t.Parallel()
	input := `{"level":"info","msg":"a"}
{"level":"error","msg":"x"}
{"level":"error","msg":"x"}
`
	out, _, code := runCLI(t, input, "summary", "--top", "2")
	if code != cmd.ExitMatched {
		t.Fatalf("exit %d, stdout: %s", code, out)
	}
	if !strings.Contains(out, "Total lines:   3") {
		t.Errorf("missing total: %s", out)
	}
	if !strings.Contains(out, "error") {
		t.Errorf("missing error count: %s", out)
	}
}

func TestUnknownCommand(t *testing.T) {
	t.Parallel()
	_, _, code := runCLI(t, "", "nonexistent")
	if code != cmd.ExitError {
		t.Errorf("got %d, want %d", code, cmd.ExitError)
	}
}
