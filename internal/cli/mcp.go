package cli

import (
	"bytes"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/bwishan/ynab-cli/internal/mcp"
)

func (a *App) registerMCPCommand() {
	mcpCmd := &cobra.Command{
		Use:   "mcp",
		Short: "Model Context Protocol server",
	}

	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the MCP server",
		Long: `Start a Model Context Protocol (MCP) server that exposes YNAB CLI
commands via two tools: search_commands and run_command.

Supports two transport modes:
  stdio (default) — reads JSON-RPC from stdin, writes to stdout
  http            — starts an HTTP server on the given port`,
		RunE: func(cmd *cobra.Command, args []string) error {
			registry := mcp.NewRegistry(a.rootCmd)
			runner := a.makeRunner()
			handlers := mcp.NewToolHandlers(registry, runner)
			srv := mcp.NewServer(registry, handlers, a.Version)

			transport, _ := cmd.Flags().GetString("transport")
			port, _ := cmd.Flags().GetInt("port")

			switch transport {
			case "stdio":
				fmt.Fprintln(cmd.ErrOrStderr(), "Starting MCP server (stdio)")
				return srv.RunStdio()
			case "http":
				addr := fmt.Sprintf(":%d", port)
				fmt.Fprintf(cmd.ErrOrStderr(), "Starting MCP server (http) on %s\n", addr)
				return srv.RunHTTP(addr)
			default:
				return fmt.Errorf("invalid transport %q (valid: stdio, http)", transport)
			}
		},
	}

	serveCmd.Flags().String("transport", "stdio", "Transport mode: stdio or http")
	serveCmd.Flags().Int("port", 8080, "HTTP port (only used with --transport http)")

	mcpCmd.AddCommand(serveCmd)
	a.rootCmd.AddCommand(mcpCmd)
}

// makeRunner returns a CommandRunner that executes commands via a fresh App instance.
func (a *App) makeRunner() mcp.CommandRunner {
	return func(args []string) (string, error) {
		app := New(a.Version, a.Commit, a.BuildDate)
		var buf bytes.Buffer
		app.SetOutput(&buf)

		err := app.Run(args)
		return buf.String(), err
	}
}
