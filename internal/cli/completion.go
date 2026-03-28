package cli

import "github.com/spf13/cobra"

func newCompletionCommand(root *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion",
		Short: "Generate shell completions",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "bash",
		Short: "Generate bash completion",
		RunE: func(cmd *cobra.Command, args []string) error {
			return root.GenBashCompletion(cmd.OutOrStdout())
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "zsh",
		Short: "Generate zsh completion",
		RunE: func(cmd *cobra.Command, args []string) error {
			return root.GenZshCompletion(cmd.OutOrStdout())
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "fish",
		Short: "Generate fish completion",
		RunE: func(cmd *cobra.Command, args []string) error {
			return root.GenFishCompletion(cmd.OutOrStdout(), true)
		},
	})

	return cmd
}
