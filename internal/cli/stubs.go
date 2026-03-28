package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newToolsCommand(state *State) *cobra.Command {
	cmd := newNotImplementedCommand(state, "tools", "List tools or inspect one tool")
	cmd.Use = "tools [server] [tool]"
	return cmd
}

func newToolCommand(state *State) *cobra.Command {
	cmd := newNotImplementedCommand(state, "tool", "Invoke a tool")
	cmd.Use = "tool [server] <tool> [args...]"
	return cmd
}

func newNotImplementedCommand(state *State, use, short string) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.Options.Invocation.IsExposedCommand() {
				server, err := state.BoundServer()
				if err != nil {
					return err
				}
				return fmt.Errorf("%s for server %q is not implemented yet", cmd.Name(), server.Name)
			}
			return fmt.Errorf("%s is not implemented yet", cmd.Name())
		},
	}
}
