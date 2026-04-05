package mcp

import (
	"testing"

	"github.com/spf13/cobra"
)

func newTestRoot() *cobra.Command {
	root := &cobra.Command{Use: "ynab"}

	plans := &cobra.Command{Use: "plans", Short: "Manage budget plans"}
	plans.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all plans",
		RunE:  func(cmd *cobra.Command, args []string) error { return nil },
	})
	getCmd := &cobra.Command{
		Use:   "get <plan-id>",
		Short: "Get a specific plan",
		RunE:  func(cmd *cobra.Command, args []string) error { return nil },
	}
	getCmd.Flags().Int64("last-knowledge", 0, "Delta sync knowledge number")
	plans.AddCommand(getCmd)

	txns := &cobra.Command{Use: "transactions", Short: "Manage transactions"}
	listCmd := &cobra.Command{
		Use:   "list <plan-id>",
		Short: "List transactions for a plan",
		RunE:  func(cmd *cobra.Command, args []string) error { return nil },
	}
	listCmd.Flags().String("since-date", "", "Only return transactions on or after this date")
	listCmd.Flags().String("type", "", "Filter by type: uncategorized or unapproved")
	txns.AddCommand(listCmd)
	txns.AddCommand(&cobra.Command{
		Use:   "create <plan-id>",
		Short: "Create a transaction",
		RunE:  func(cmd *cobra.Command, args []string) error { return nil },
	})
	txns.AddCommand(&cobra.Command{
		Use:   "delete <plan-id> <transaction-id>",
		Short: "Delete a transaction",
		RunE:  func(cmd *cobra.Command, args []string) error { return nil },
	})

	accounts := &cobra.Command{Use: "accounts", Short: "Manage accounts"}
	accounts.AddCommand(&cobra.Command{
		Use:   "list <plan-id>",
		Short: "List all accounts",
		RunE:  func(cmd *cobra.Command, args []string) error { return nil },
	})

	// Commands that should be excluded
	root.AddCommand(&cobra.Command{Use: "configure", Short: "Save config", RunE: func(cmd *cobra.Command, args []string) error { return nil }})
	mcpCmd := &cobra.Command{Use: "mcp", Short: "MCP server"}
	mcpCmd.AddCommand(&cobra.Command{Use: "serve", Short: "Start MCP server", RunE: func(cmd *cobra.Command, args []string) error { return nil }})
	root.AddCommand(mcpCmd)

	root.AddCommand(plans, txns, accounts)
	return root
}

func TestNewRegistry(t *testing.T) {
	reg := NewRegistry(newTestRoot())

	cmds := reg.Commands()
	if len(cmds) != 6 {
		t.Fatalf("expected 6 commands, got %d: %v", len(cmds), cmdNames(cmds))
	}

	// Verify configure and mcp serve are excluded
	for _, cmd := range cmds {
		if cmd.Name == "configure" || cmd.Name == "mcp serve" {
			t.Errorf("should have excluded %q", cmd.Name)
		}
	}
}

func TestRegistrySearchByResource(t *testing.T) {
	reg := NewRegistry(newTestRoot())

	results := reg.Search("", "transactions")
	if len(results) != 3 {
		t.Fatalf("expected 3 transaction commands, got %d", len(results))
	}
	for _, r := range results {
		if r.Resource != "transactions" {
			t.Errorf("expected resource 'transactions', got %q", r.Resource)
		}
	}
}

func TestRegistrySearchByQuery(t *testing.T) {
	reg := NewRegistry(newTestRoot())

	results := reg.Search("delete", "")
	if len(results) != 1 {
		t.Fatalf("expected 1 result for 'delete', got %d", len(results))
	}
	if results[0].Name != "transactions delete" {
		t.Errorf("expected 'transactions delete', got %q", results[0].Name)
	}
}

func TestRegistrySearchByQueryAndResource(t *testing.T) {
	reg := NewRegistry(newTestRoot())

	results := reg.Search("list", "transactions")
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Name != "transactions list" {
		t.Errorf("expected 'transactions list', got %q", results[0].Name)
	}
}

func TestRegistrySearchEmpty(t *testing.T) {
	reg := NewRegistry(newTestRoot())

	results := reg.Search("", "")
	if len(results) != 6 {
		t.Fatalf("expected 6 commands for empty search, got %d", len(results))
	}
}

func TestRegistrySearchByFlag(t *testing.T) {
	reg := NewRegistry(newTestRoot())

	results := reg.Search("since-date", "")
	if len(results) != 1 {
		t.Fatalf("expected 1 result for 'since-date', got %d", len(results))
	}
	if results[0].Name != "transactions list" {
		t.Errorf("expected 'transactions list', got %q", results[0].Name)
	}
}

func TestRegistryResources(t *testing.T) {
	reg := NewRegistry(newTestRoot())

	resources := reg.Resources()
	if len(resources) != 3 {
		t.Fatalf("expected 3 resources, got %d: %v", len(resources), resources)
	}
}

func TestParseArgs(t *testing.T) {
	tests := []struct {
		use  string
		want []ArgInfo
	}{
		{"list", nil},
		{"get <plan-id>", []ArgInfo{{Name: "plan-id", Required: true}}},
		{"list [plan-id]", []ArgInfo{{Name: "plan-id", Required: false}}},
		{"delete <plan-id> <transaction-id>", []ArgInfo{
			{Name: "plan-id", Required: true},
			{Name: "transaction-id", Required: true},
		}},
	}

	for _, tt := range tests {
		got := parseArgs(tt.use)
		if len(got) != len(tt.want) {
			t.Errorf("parseArgs(%q): got %d args, want %d", tt.use, len(got), len(tt.want))
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("parseArgs(%q)[%d]: got %+v, want %+v", tt.use, i, got[i], tt.want[i])
			}
		}
	}
}

func cmdNames(cmds []CommandInfo) []string {
	var names []string
	for _, c := range cmds {
		names = append(names, c.Name)
	}
	return names
}
