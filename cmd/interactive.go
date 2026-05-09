package cmd

import (
	"github.com/spf13/cobra"
)

func newInteractiveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "interactive",
		Short: "Interactive log viewer",
		Long:  "Launch an interactive TUI for filtering and viewing logs.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return interactiveRun(cmd, args)
		},
	}
	return cmd
}