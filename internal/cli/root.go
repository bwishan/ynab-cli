// Package cli implements the command-line interface for ynab-cli.
package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/bwishan/ynab-cli/internal/api"
	"github.com/bwishan/ynab-cli/internal/config"
	"github.com/bwishan/ynab-cli/internal/output"
)

// App is the root CLI application.
type App struct {
	rootCmd   *cobra.Command
	Version   string
	Commit    string
	BuildDate string
	client    *api.Client
	printer   *output.Printer
	cfg       *config.Config
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
  ynab transactions sync <plan-id>
  ynab categories list <plan-id> --output table`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Skip auth for commands that don't need it
			if cmd.RunE == nil && cmd.Run == nil {
				return nil
			}
			if cmd.Name() == "configure" {
				return nil
			}
			// Local-only commands need printer+config but not an API token.
			if cmd.Annotations["localOnly"] == "true" {
				return a.initPrinterAndConfig(cmd)
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

	return a
}

func (a *App) initContext(cmd *cobra.Command) error {
	tokenFlag, _ := cmd.Flags().GetString("token")
	resolvedToken, err := config.ResolveToken(tokenFlag)
	if err != nil {
		return err
	}
	a.client = api.NewClient(resolvedToken)
	return a.initPrinterAndConfig(cmd)
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
		Long:  "Save access token, default plan, and sync settings to ~/.config/ynab-cli/config.json",
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

			syncEnable, _ := cmd.Flags().GetBool("transaction-sync")
			syncDisable, _ := cmd.Flags().GetBool("transaction-sync-off")
			syncDB, _ := cmd.Flags().GetString("transaction-sync-db")
			if syncEnable && syncDisable {
				return fmt.Errorf("cannot use --transaction-sync and --transaction-sync-off together")
			}
			if syncEnable {
				cfg.TransactionSync = true
			}
			if syncDisable {
				cfg.TransactionSync = false
			}
			if syncDB != "" {
				cfg.TransactionSyncDB = syncDB
			}
			if cfg.TransactionSync && cfg.TransactionSyncDB == "" {
				defaultDB, err := config.DefaultSyncDBPath()
				if err != nil {
					return err
				}
				cfg.TransactionSyncDB = defaultDB
			}

			if err := config.Save(cfg); err != nil {
				return err
			}
			fmt.Println("Configuration saved.")
			return nil
		},
	}

	configureCmd.Flags().String("default-plan", "", "Default plan ID")
	configureCmd.Flags().Bool("transaction-sync", false, "Enable remembered transaction sync/cache mode")
	configureCmd.Flags().Bool("transaction-sync-off", false, "Disable remembered transaction sync/cache mode")
	configureCmd.Flags().String("transaction-sync-db", "", "SQLite DB path for remembered transaction sync/cache")

	a.rootCmd.AddCommand(configureCmd)
}
