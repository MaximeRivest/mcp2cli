package cli

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/maximerivest/mcp2cli/internal/config"
	"github.com/maximerivest/mcp2cli/internal/expose"
	"github.com/spf13/cobra"
)

func newAddCommand(state *State) *cobra.Command {
	var (
		command   string
		url       string
		auth      string
		bearerEnv string
		local     bool
		roots     []string
	)

	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Register a server",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name, err := config.NormalizeCommandName(args[0])
			if err != nil {
				return err
			}
			if command == "" && url == "" {
				return fmt.Errorf("one of --command or --url is required")
			}
			if command != "" && url != "" {
				return fmt.Errorf("--command and --url are mutually exclusive")
			}

			repo, err := state.Repo()
			if err != nil {
				return err
			}

			scope := config.SourceGlobal
			if local {
				scope = config.SourceLocal
			}

			server := &config.Server{
				Command:   strings.TrimSpace(command),
				URL:       strings.TrimSpace(url),
				Auth:      strings.TrimSpace(auth),
				BearerEnv: strings.TrimSpace(bearerEnv),
				Roots:     append([]string(nil), roots...),
			}
			if err := repo.UpsertServer(scope, name, server); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "added server %q to %s config\n", name, scope)
			return nil
		},
	}

	cmd.Flags().StringVar(&command, "command", "", "Local server command to run")
	cmd.Flags().StringVar(&url, "url", "", "Remote MCP server URL")
	cmd.Flags().StringVar(&auth, "auth", "", "Authentication mode")
	cmd.Flags().StringVar(&bearerEnv, "bearer-env", "", "Environment variable containing a bearer token")
	cmd.Flags().StringSliceVar(&roots, "root", nil, "Root path exposed to the server (repeatable)")
	cmd.Flags().BoolVar(&local, "local", false, "Write to .mcp2cli.yaml instead of global config")
	return cmd
}

func newListCommand(state *State) *cobra.Command {
	return &cobra.Command{
		Use:   "ls",
		Short: "List registered servers",
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, err := state.Repo()
			if err != nil {
				return err
			}
			servers, err := repo.ListServers()
			if err != nil {
				return err
			}
			if len(servers) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "no registered servers")
				return nil
			}

			writer := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(writer, "NAME\tSOURCE\tTYPE\tEXPOSED\tTARGET")
			for _, server := range servers {
				exposed := "-"
				if len(server.ExposeAs) > 0 {
					exposed = strings.Join(server.ExposeAs, ",")
				}
				typeName := "command"
				target := server.Command
				if server.URL != "" {
					typeName = "url"
					target = server.URL
				}
				fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n", server.Name, server.Source, typeName, exposed, target)
			}
			return writer.Flush()
		},
	}
}

func newRemoveCommand(state *State) *cobra.Command {
	return &cobra.Command{
		Use:   "rm <name>",
		Short: "Remove a registered server",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, err := state.Repo()
			if err != nil {
				return err
			}
			server, err := repo.ResolveServer(args[0])
			if err != nil {
				return err
			}
			for _, exposedName := range server.ExposeAs {
				if err := expose.Remove(repo.Paths.ExposeBinDir, exposedName); err != nil {
					return err
				}
			}
			if err := repo.RemoveServer(server.Source, server.Name); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "removed server %q from %s config\n", server.Name, server.Source)
			return nil
		},
	}
}

func newExposeCommand(state *State) *cobra.Command {
	var as string

	cmd := &cobra.Command{
		Use:   "expose <server>",
		Short: "Create an exposed command for a server",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, err := state.Repo()
			if err != nil {
				return err
			}
			server, err := repo.ResolveServer(args[0])
			if err != nil {
				return err
			}

			exposedName := strings.TrimSpace(as)
			if exposedName == "" {
				exposedName, err = config.DefaultExposeName(server.Name)
				if err != nil {
					return err
				}
			} else {
				exposedName, err = config.NormalizeCommandName(exposedName)
				if err != nil {
					return err
				}
			}

			if err := repo.AddExpose(server.Source, server.Name, exposedName); err != nil {
				return err
			}
			executable, err := os.Executable()
			if err != nil {
				return fmt.Errorf("find current executable: %w", err)
			}
			shimPath, err := expose.Create(repo.Paths.ExposeBinDir, exposedName, executable)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "exposed %q as %q\n", server.Name, exposedName)
			fmt.Fprintf(cmd.OutOrStdout(), "shim: %s\n", shimPath)
			fmt.Fprintf(cmd.OutOrStdout(), "add %s to PATH if needed\n", repo.Paths.ExposeBinDir)
			return nil
		},
	}

	cmd.Flags().StringVar(&as, "as", "", "Full exposed command name to create")
	return cmd
}

func newUnexposeCommand(state *State) *cobra.Command {
	var as string

	cmd := &cobra.Command{
		Use:   "unexpose <server>",
		Short: "Remove an exposed command",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, err := state.Repo()
			if err != nil {
				return err
			}
			server, err := repo.ResolveServer(args[0])
			if err != nil {
				return err
			}

			exposedName := strings.TrimSpace(as)
			if exposedName == "" {
				exposedName, err = config.DefaultExposeName(server.Name)
				if err != nil {
					return err
				}
			} else {
				exposedName, err = config.NormalizeCommandName(exposedName)
				if err != nil {
					return err
				}
			}

			if err := repo.RemoveExpose(server.Source, server.Name, exposedName); err != nil {
				return err
			}
			if err := expose.Remove(repo.Paths.ExposeBinDir, exposedName); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "removed exposed command %q for server %q\n", exposedName, server.Name)
			return nil
		},
	}

	cmd.Flags().StringVar(&as, "as", "", "Full exposed command name to remove")
	return cmd
}
