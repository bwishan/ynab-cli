// Package config handles configuration for the ynab CLI.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds the CLI configuration.
type Config struct {
	AccessToken       string `json:"access_token"`
	DefaultPlan       string `json:"default_plan,omitempty"`
	TransactionSync   bool   `json:"transaction_sync,omitempty"`
	TransactionSyncDB string `json:"transaction_sync_db,omitempty"`
}

// configDir returns the config directory path.
func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "ynab-cli"), nil
}

// configPath returns the config file path.
func configPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// DefaultSyncDBPath returns the default SQLite path used for transaction sync.
func DefaultSyncDBPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "transactions.db"), nil
}

// Load reads the config from disk. Returns an empty config if the file doesn't exist.
func Load() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return &Config{}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	return &cfg, nil
}

// Save writes the config to disk.
func Save(cfg *Config) error {
	dir, err := configDir()
	if err != nil {
		return fmt.Errorf("getting config dir: %w", err)
	}

	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	return nil
}

// ResolveToken returns the access token from (in priority order):
// 1. The --token flag value
// 2. YNAB_ACCESS_TOKEN env var
// 3. Config file
func ResolveToken(flagValue string) (string, error) {
	if flagValue != "" {
		return flagValue, nil
	}
	if token := os.Getenv("YNAB_ACCESS_TOKEN"); token != "" {
		return token, nil
	}
	cfg, err := Load()
	if err != nil {
		return "", err
	}
	if cfg.AccessToken != "" {
		return cfg.AccessToken, nil
	}
	return "", fmt.Errorf("no access token found. Set YNAB_ACCESS_TOKEN, use --token, or run: ynab configure --token <token>")
}

// ResolvePlan returns the plan ID from (in priority order):
// 1. The explicit argument
// 2. YNAB_PLAN_ID env var
// 3. Config file default_plan
func ResolvePlan(explicit string) (string, error) {
	if explicit != "" {
		return explicit, nil
	}
	if planID := os.Getenv("YNAB_PLAN_ID"); planID != "" {
		return planID, nil
	}
	cfg, err := Load()
	if err != nil {
		return "", err
	}
	if cfg.DefaultPlan != "" {
		return cfg.DefaultPlan, nil
	}
	return "", fmt.Errorf("no plan ID specified. Provide <plan-id> argument, set YNAB_PLAN_ID, or run: ynab configure --default-plan <id>")
}
