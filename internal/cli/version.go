package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newVersionCommand(state *State) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(cmd.OutOrStdout(), "version: %s\n", state.Options.Version)
			if state.Options.Commit != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "commit: %s\n", state.Options.Commit)
			}
			if state.Options.BuildDate != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "built: %s\n", state.Options.BuildDate)
			}
			return nil
		},
	}
}
