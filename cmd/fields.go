package cmd

import (
	"errors"

	"github.com/spf13/cobra"
)

func newFieldsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fields [file...]",
		Short: "Inspect available fields in a log file",
		Long:  "List unique field names with their occurrence counts. Useful before crafting --field filters.",
		RunE: func(_ *cobra.Command, _ []string) error {
			return errors.New("fields: not yet implemented (planned for v0.2)")
		},
	}
	return cmd
}
