package cmd

import (
	"errors"

	"github.com/spf13/cobra"
)

func newTailCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tail [flags] file",
		Short: "Watch a log file in real time",
		Long:  "Tail a file like `tail -f` and apply filters to incoming lines.",
		RunE: func(_ *cobra.Command, _ []string) error {
			// Implementation deferred to a follow-up commit. We're shipping the
			// command surface and config plumbing first; tail uses fsnotify for
			// rotation/truncation handling and warrants its own focused PR.
			return errors.New("tail: not yet implemented (planned for v0.2)")
		},
	}
	return cmd
}
