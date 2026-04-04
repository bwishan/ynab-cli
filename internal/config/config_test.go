package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveToken_FlagTakesPriority(t *testing.T) {
	t.Setenv("YNAB_ACCESS_TOKEN", "env-token")

	got, err := ResolveToken("flag-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "flag-token" {
		t.Errorf("expected 'flag-token', got %q", got)
	}
}

func TestResolveToken_EnvFallback(t *testing.T) {
	t.Setenv("YNAB_ACCESS_TOKEN", "env-token")

	got, err := ResolveToken("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "env-token" {
		t.Errorf("expected 'env-token', got %q", got)
	}
}

func TestResolveToken_ConfigFileFallback(t *testing.T) {
	t.Setenv("YNAB_ACCESS_TOKEN", "")

	// Set HOME to a temp dir with a config file
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	dir := filepath.Join(tmp, ".config", "ynab-cli")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte(`{"access_token":"file-token"}`), 0600); err != nil {
		t.Fatal(err)
	}

	got, err := ResolveToken("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "file-token" {
		t.Errorf("expected 'file-token', got %q", got)
	}
}

func TestResolveToken_ErrorWhenNotFound(t *testing.T) {
	t.Setenv("YNAB_ACCESS_TOKEN", "")
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	_, err := ResolveToken("")
	if err == nil {
		t.Fatal("expected error when no token found")
	}
}

func TestResolvePlan_ExplicitTakesPriority(t *testing.T) {
	t.Setenv("YNAB_PLAN_ID", "env-plan")

	got, err := ResolvePlan("explicit-plan")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "explicit-plan" {
		t.Errorf("expected 'explicit-plan', got %q", got)
	}
}

func TestResolvePlan_EnvFallback(t *testing.T) {
	t.Setenv("YNAB_PLAN_ID", "env-plan")

	got, err := ResolvePlan("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "env-plan" {
		t.Errorf("expected 'env-plan', got %q", got)
	}
}

func TestResolvePlan_ConfigFileFallback(t *testing.T) {
	t.Setenv("YNAB_PLAN_ID", "")
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	dir := filepath.Join(tmp, ".config", "ynab-cli")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte(`{"default_plan":"file-plan"}`), 0600); err != nil {
		t.Fatal(err)
	}

	got, err := ResolvePlan("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "file-plan" {
		t.Errorf("expected 'file-plan', got %q", got)
	}
}

func TestResolvePlan_ErrorWhenNotFound(t *testing.T) {
	t.Setenv("YNAB_PLAN_ID", "")
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	_, err := ResolvePlan("")
	if err == nil {
		t.Fatal("expected error when no plan found")
	}
}

func TestSaveAndLoad_RoundTrip(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	cfg := &Config{
		AccessToken:       "my-token-123",
		DefaultPlan:       "plan-abc",
		TransactionSync:   true,
		TransactionSyncDB: "/tmp/ynab-sync.db",
	}

	if err := Save(cfg); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if loaded.AccessToken != cfg.AccessToken {
		t.Errorf("AccessToken: expected %q, got %q", cfg.AccessToken, loaded.AccessToken)
	}
	if loaded.DefaultPlan != cfg.DefaultPlan {
		t.Errorf("DefaultPlan: expected %q, got %q", cfg.DefaultPlan, loaded.DefaultPlan)
	}
	if loaded.TransactionSync != cfg.TransactionSync {
		t.Errorf("TransactionSync: expected %v, got %v", cfg.TransactionSync, loaded.TransactionSync)
	}
	if loaded.TransactionSyncDB != cfg.TransactionSyncDB {
		t.Errorf("TransactionSyncDB: expected %q, got %q", cfg.TransactionSyncDB, loaded.TransactionSyncDB)
	}
}

func TestLoad_ReturnsEmptyWhenFileDoesNotExist(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.AccessToken != "" || cfg.DefaultPlan != "" || cfg.TransactionSync || cfg.TransactionSyncDB != "" {
		t.Errorf("expected empty config, got: %+v", cfg)
	}
}

func TestDefaultSyncDBPath(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	path, err := DefaultSyncDBPath()
	if err != nil {
		t.Fatalf("DefaultSyncDBPath returned error: %v", err)
	}
	if path == "" {
		t.Error("expected non-empty default sync DB path")
	}
	// Should be inside the ynab-cli config dir.
	if !containsPath(path, "ynab-cli") {
		t.Errorf("expected path inside ynab-cli dir, got: %q", path)
	}
}

func TestDefaultSyncDBPath_FilenameIsTransactionsDB(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	path, err := DefaultSyncDBPath()
	if err != nil {
		t.Fatalf("DefaultSyncDBPath returned error: %v", err)
	}
	if filepath.Base(path) != "transactions.db" {
		t.Errorf("expected filename transactions.db, got %q", filepath.Base(path))
	}
}

func containsPath(path, segment string) bool {
	for _, part := range splitPath(path) {
		if part == segment {
			return true
		}
	}
	return false
}

func splitPath(path string) []string {
	var parts []string
	for path != "" {
		dir, file := splitLast(path)
		if file != "" {
			parts = append(parts, file)
		}
		if dir == path {
			break
		}
		path = dir
	}
	return parts
}

func splitLast(path string) (string, string) {
	i := len(path) - 1
	for i >= 0 && path[i] == '/' {
		i--
	}
	j := i
	for j >= 0 && path[j] != '/' {
		j--
	}
	return path[:j+1], path[j+1 : i+1]
}
