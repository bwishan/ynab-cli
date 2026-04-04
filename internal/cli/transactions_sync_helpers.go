package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/bwishan/ynab-cli/internal/config"
)

func (a *App) resolveTransactionSyncSettings(cmd *cobra.Command) (bool, string, error) {
	if a.cfg == nil {
		cfg, err := config.Load()
		if err != nil {
			return false, "", err
		}
		a.cfg = cfg
	}

	enabled := a.cfg.TransactionSync
	if cmd != nil {
		if cmd.Flags().Changed("transaction-sync") {
			flagValue, _ := cmd.Flags().GetBool("transaction-sync")
			enabled = flagValue
		}
		if cmd.Flags().Changed("transaction-sync-off") {
			flagValue, _ := cmd.Flags().GetBool("transaction-sync-off")
			if flagValue {
				enabled = false
			}
		}
	}

	dbPath := a.cfg.TransactionSyncDB
	if cmd != nil && cmd.Flags().Changed("transaction-sync-db") {
		flagValue, _ := cmd.Flags().GetString("transaction-sync-db")
		dbPath = flagValue
	}
	if enabled && dbPath == "" {
		defaultDB, err := config.DefaultSyncDBPath()
		if err != nil {
			return false, "", err
		}
		dbPath = defaultDB
	}
	return enabled, dbPath, nil
}

func (a *App) requireTransactionSyncDB(cmd *cobra.Command) (string, error) {
	_, dbPath, err := a.resolveTransactionSyncSettings(cmd)
	if err != nil {
		return "", err
	}
	if dbPath == "" {
		defaultDB, err := config.DefaultSyncDBPath()
		if err != nil {
			return "", err
		}
		return defaultDB, nil
	}
	return dbPath, nil
}

func (a *App) addTransactionSyncFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("transaction-sync", false, "Enable transaction sync/cache mode for this command")
	cmd.Flags().Bool("transaction-sync-off", false, "Disable transaction sync/cache mode for this command")
	cmd.Flags().String("transaction-sync-db", "", "SQLite DB path for transaction sync/cache")
}

func (a *App) validateTransactionSyncFlags(cmd *cobra.Command) error {
	on, _ := cmd.Flags().GetBool("transaction-sync")
	off, _ := cmd.Flags().GetBool("transaction-sync-off")
	if on && off {
		return fmt.Errorf("cannot use --transaction-sync and --transaction-sync-off together")
	}
	return nil
}
