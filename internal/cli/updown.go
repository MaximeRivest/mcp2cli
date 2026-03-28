package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/adrg/xdg"
	"github.com/maximerivest/mcp2cli/internal/daemon"
	"github.com/maximerivest/mcp2cli/internal/exitcode"
	"github.com/spf13/cobra"
)

func newUpCommand(state *State) *cobra.Command {
	return &cobra.Command{
		Use:   "up",
		Short: "Start the server in the background for fast repeated use",
		RunE: func(cmd *cobra.Command, args []string) error {
			server, err := state.BoundServer()
			if err != nil || server == nil {
				return exitcode.New(exitcode.Usage, "up requires a server context (use: mcp2cli <server> up)")
			}
			if server.Command == "" {
				return exitcode.New(exitcode.Config, "up only works with local stdio servers")
			}
			if daemon.IsRunning(xdg.DataHome, server.Name) {
				fmt.Fprintf(cmd.OutOrStdout(), "%s is already running\n", server.Name)
				return nil
			}

			// Start a detached child process running the daemon
			self, err := os.Executable()
			if err != nil {
				return exitcode.Wrap(exitcode.Internal, err, "find executable")
			}
			child := exec.Command(self, "__daemon", server.Name, server.Command)
			child.Stdout = nil
			child.Stderr = nil
			child.Stdin = nil
			child.SysProcAttr = daemonSysProcAttr()
			if err := child.Start(); err != nil {
				return exitcode.Wrap(exitcode.Transport, err, "start daemon")
			}

			// Wait briefly for the socket to appear
			for i := 0; i < 30; i++ {
				time.Sleep(100 * time.Millisecond)
				if daemon.IsRunning(xdg.DataHome, server.Name) {
					fmt.Fprintf(cmd.OutOrStdout(), "%s is running (pid %d)\n", server.Name, child.Process.Pid)
					return nil
				}
			}
			return exitcode.New(exitcode.Transport, "daemon did not start in time")
		},
	}
}

func newDownCommand(state *State) *cobra.Command {
	return &cobra.Command{
		Use:   "down",
		Short: "Stop the background server",
		RunE: func(cmd *cobra.Command, args []string) error {
			server, err := state.BoundServer()
			if err != nil || server == nil {
				return exitcode.New(exitcode.Usage, "down requires a server context (use: mcp2cli <server> down)")
			}
			if err := daemon.Stop(xdg.DataHome, server.Name); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s stopped\n", server.Name)
			return nil
		},
	}
}

func newDaemonCommand(state *State) *cobra.Command {
	return &cobra.Command{
		Use:    "__daemon",
		Hidden: true,
		Args:   cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			serverName := args[0]
			command := args[1]
			return daemon.Run(context.Background(), command, xdg.DataHome, serverName)
		},
	}
}
