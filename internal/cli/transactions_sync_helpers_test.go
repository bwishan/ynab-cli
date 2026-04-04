package cli

import (
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"

	"github.com/bwishan/ynab-cli/internal/config"
)

func TestResolveTransactionSyncSettings_UsesConfig(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	cfg := &config.Config{TransactionSync: true, TransactionSyncDB: "/tmp/ynab-sync.db"}
	if err := config.Save(cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	app := &App{}
	enabled, dbPath, err := app.resolveTransactionSyncSettings(nil)
	if err != nil {
		t.Fatalf("resolveTransactionSyncSettings: %v", err)
	}
	if !enabled {
		t.Fatal("expected sync to be enabled")
	}
	if dbPath != "/tmp/ynab-sync.db" {
		t.Fatalf("dbPath = %q, want /tmp/ynab-sync.db", dbPath)
	}
}

func TestResolveTransactionSyncSettings_DefaultDBWhenEnabled(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	if err := config.Save(&config.Config{TransactionSync: true}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	app := &App{}
	enabled, dbPath, err := app.resolveTransactionSyncSettings(nil)
	if err != nil {
		t.Fatalf("resolveTransactionSyncSettings: %v", err)
	}
	if !enabled {
		t.Fatal("expected sync to be enabled")
	}
	want := filepath.Join(t.TempDir(), ".config", "ynab-cli", "transactions.db")
	_ = want // tempdir changes per call; assert suffix instead
	if filepath.Base(dbPath) != "transactions.db" {
		t.Fatalf("dbPath = %q, want basename transactions.db", dbPath)
	}
}

func TestResolveTransactionSyncSettings_CommandFlagsOverrideConfig(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	if err := config.Save(&config.Config{TransactionSync: false, TransactionSyncDB: "/tmp/from-config.db"}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	cmd := &cobra.Command{Use: "test"}
	app := &App{}
	app.addTransactionSyncFlags(cmd)
	if err := cmd.Flags().Set("transaction-sync", "true"); err != nil {
		t.Fatalf("set transaction-sync: %v", err)
	}
	if err := cmd.Flags().Set("transaction-sync-db", "/tmp/from-flag.db"); err != nil {
		t.Fatalf("set transaction-sync-db: %v", err)
	}

	enabled, dbPath, err := app.resolveTransactionSyncSettings(cmd)
	if err != nil {
		t.Fatalf("resolveTransactionSyncSettings: %v", err)
	}
	if !enabled {
		t.Fatal("expected sync to be enabled by flag")
	}
	if dbPath != "/tmp/from-flag.db" {
		t.Fatalf("dbPath = %q, want /tmp/from-flag.db", dbPath)
	}
}

func TestRequireTransactionSyncDB_ReturnsDefaultPath(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	app := &App{cfg: &config.Config{}}
	got, err := app.requireTransactionSyncDB(nil)
	if err != nil {
		t.Fatalf("requireTransactionSyncDB: %v", err)
	}
	if filepath.Base(got) != "transactions.db" {
		t.Fatalf("got %q, want basename transactions.db", got)
	}
}
