package mcp

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func newTestHandlers() *ToolHandlers {
	reg := NewRegistry(newTestRoot())
	runner := func(args []string) (string, error) {
		return `{"data":{"plans":[]}}`, nil
	}
	return NewToolHandlers(reg, runner)
}

func TestDefinitionsCount(t *testing.T) {
	h := newTestHandlers()
	defs := h.Definitions()
	if len(defs) != 2 {
		t.Fatalf("expected 2 tool definitions, got %d", len(defs))
	}
	if defs[0].Name != "search_commands" {
		t.Errorf("expected first tool 'search_commands', got %q", defs[0].Name)
	}
	if defs[1].Name != "run_command" {
		t.Errorf("expected second tool 'run_command', got %q", defs[1].Name)
	}
}

func TestSearchDescriptionDynamic(t *testing.T) {
	h := newTestHandlers()
	defs := h.Definitions()
	desc := defs[0].Description
	// Should contain all resources from the test tree
	for _, res := range []string{"plans", "transactions", "accounts"} {
		if !strings.Contains(desc, res) {
			t.Errorf("search description missing resource %q: %s", res, desc)
		}
	}
}

func TestHandleSearchAll(t *testing.T) {
	h := newTestHandlers()
	result, err := h.Call("search_commands", json.RawMessage(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Content) != 1 {
		t.Fatal("expected 1 content item")
	}
	var cmds []CommandInfo
	if err := json.Unmarshal([]byte(result.Content[0].Text), &cmds); err != nil {
		t.Fatalf("failed to parse search result: %v", err)
	}
	if len(cmds) != 6 {
		t.Errorf("expected 6 commands, got %d", len(cmds))
	}
}

func TestHandleSearchByResource(t *testing.T) {
	h := newTestHandlers()
	result, err := h.Call("search_commands", json.RawMessage(`{"resource":"plans"}`))
	if err != nil {
		t.Fatal(err)
	}
	var cmds []CommandInfo
	json.Unmarshal([]byte(result.Content[0].Text), &cmds)
	if len(cmds) != 2 {
		t.Errorf("expected 2 plan commands, got %d", len(cmds))
	}
}

func TestHandleRunSuccess(t *testing.T) {
	h := newTestHandlers()
	result, err := h.Call("run_command", json.RawMessage(`{"command":"plans list"}`))
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Error("expected success, got error")
	}
	if !strings.Contains(result.Content[0].Text, "plans") {
		t.Errorf("unexpected output: %s", result.Content[0].Text)
	}
}

func TestHandleRunError(t *testing.T) {
	reg := NewRegistry(newTestRoot())
	runner := func(args []string) (string, error) {
		return "", fmt.Errorf("token required")
	}
	h := NewToolHandlers(reg, runner)

	result, err := h.Call("run_command", json.RawMessage(`{"command":"plans list"}`))
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error result")
	}
	if !strings.Contains(result.Content[0].Text, "token required") {
		t.Errorf("expected error message, got: %s", result.Content[0].Text)
	}
}

func TestHandleRunMissingCommand(t *testing.T) {
	h := newTestHandlers()
	_, err := h.Call("run_command", json.RawMessage(`{}`))
	if err == nil {
		t.Error("expected error for missing command")
	}
}

func TestHandleUnknownTool(t *testing.T) {
	h := newTestHandlers()
	_, err := h.Call("nonexistent", json.RawMessage(`{}`))
	if err == nil {
		t.Error("expected error for unknown tool")
	}
}

func TestTokenize(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"plans list", []string{"plans", "list"}},
		{`transactions create abc --memo "Morning coffee"`, []string{"transactions", "create", "abc", "--memo", "Morning coffee"}},
		{"", nil},
		{"  plans   list  ", []string{"plans", "list"}},
	}

	for _, tt := range tests {
		got := tokenize(tt.input)
		if len(got) != len(tt.want) {
			t.Errorf("tokenize(%q): got %v, want %v", tt.input, got, tt.want)
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("tokenize(%q)[%d]: got %q, want %q", tt.input, i, got[i], tt.want[i])
			}
		}
	}
}
