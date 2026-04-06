// Package cli implements the command-line interface for ynab-cli.
package cli

import (
"fmt"
"io"

"github.com/spf13/cobra"

"github.com/bwishan/ynab-cli/internal/api"
"github.com/bwishan/ynab-cli/internal/config"
"github.com/bwishan/ynab-cli/internal/output"
)

// App is the root CLI application.
type App struct {
rootCmd      *cobra.Command
Version      string
Commit       string
BuildDate    string
client       *api.Client
printer      *output.Printer
outputWriter io.Writer
}

// New creates the root CLI app and registers all commands.
func New(version, commit, buildDate string) *App {
a := &App{
Version:   version,
Commit:    commit,
BuildDate: buildDate,
}

rootCmd := &cobra.Command{
Use:   "ynab",
Short: "Agent-friendly CLI for the YNAB API",
Long: `ynab - Agent-friendly CLI for the YNAB API

ENVIRONMENT VARIABLES:
  YNAB_ACCESS_TOKEN    Access token for authentication
  YNAB_PLAN_ID         Default plan ID

EXAMPLES:
  ynab plans list
  ynab accounts list <plan-id>
  ynab transactions list <plan-id> --since-date 2024-01-01
  ynab categories list <plan-id> --output table`,
SilenceUsage:  true,
SilenceErrors: true,
PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
// Skip auth for commands that don't need it
if cmd.RunE == nil && cmd.Run == nil {
return nil
}
if cmd.Name() == "configure" || cmd.Name() == "serve" {
return nil
}
return a.initContext(cmd)
},
}

rootCmd.PersistentFlags().String("token", "", "YNAB access token (or set YNAB_ACCESS_TOKEN)")
rootCmd.PersistentFlags().String("output", "json", "Output format: json, table, csv")
rootCmd.PersistentFlags().Bool("pretty", false, "Pretty-print JSON output")

rootCmd.Version = fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, buildDate)
rootCmd.SetVersionTemplate("ynab {{.Version}}\n")

a.rootCmd = rootCmd

a.registerConfigureCommand()
a.registerUserCommands()
a.registerPlanCommands()
a.registerAccountCommands()
a.registerCategoryCommands()
a.registerTransactionCommands()
a.registerPayeeCommands()
a.registerScheduledCommands()
a.registerMonthCommands()
a.registerMoneyMovementCommands()
a.registerMCPCommand()

return a
}

func (a *App) initContext(cmd *cobra.Command) error {
tokenFlag, _ := cmd.Flags().GetString("token")
resolvedToken, err := config.ResolveToken(tokenFlag)
if err != nil {
return err
}

a.client = api.NewClient(resolvedToken)

formatStr, _ := cmd.Flags().GetString("output")
format := output.Format(formatStr)
switch format {
case output.FormatJSON, output.FormatTable, output.FormatCSV:
default:
return fmt.Errorf("invalid output format %q (valid: json, table, csv)", formatStr)
}

pretty, _ := cmd.Flags().GetBool("pretty")
if a.outputWriter != nil {
a.printer = output.NewPrinterWithWriter(format, pretty, a.outputWriter)
} else {
a.printer = output.NewPrinter(format, pretty)
}

return nil
}

// SetOutput redirects all command output to the given writer instead of stdout.
func (a *App) SetOutput(w io.Writer) {
a.outputWriter = w
}

// RootCmd returns the root cobra command for introspection.
func (a *App) RootCmd() *cobra.Command {
return a.rootCmd
}

// Run parses the CLI arguments and dispatches to the appropriate handler.
func (a *App) Run(args []string) error {
a.rootCmd.SetArgs(args)
return a.rootCmd.Execute()
}

func (a *App) registerConfigureCommand() {
configureCmd := &cobra.Command{
Use:   "configure",
Short: "Save CLI configuration",
Long:  "Save access token and default plan to ~/.config/ynab-cli/config.json",
RunE: func(cmd *cobra.Command, args []string) error {
cfg, err := config.Load()
if err != nil {
cfg = &config.Config{}
}

token, _ := cmd.Flags().GetString("token")
if token != "" {
cfg.AccessToken = token
}

defaultPlan, _ := cmd.Flags().GetString("default-plan")
if defaultPlan != "" {
cfg.DefaultPlan = defaultPlan
}

if err := config.Save(cfg); err != nil {
return err
}
fmt.Println("Configuration saved.")
return nil
},
}

configureCmd.Flags().String("default-plan", "", "Default plan ID")

a.rootCmd.AddCommand(configureCmd)
}
