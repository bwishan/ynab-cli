package mcp

import (
	"encoding/json"
	"fmt"
	"strings"
)

// CommandRunner executes a CLI command with the given args and returns the output.
type CommandRunner func(args []string) (string, error)

// ToolHandlers manages the MCP tool definitions and execution.
type ToolHandlers struct {
	registry *Registry
	runner   CommandRunner
}

// NewToolHandlers creates handlers backed by the given registry and command runner.
func NewToolHandlers(registry *Registry, runner CommandRunner) *ToolHandlers {
	return &ToolHandlers{
		registry: registry,
		runner:   runner,
	}
}

// Definitions returns the MCP tool definitions for tools/list.
func (h *ToolHandlers) Definitions() []ToolDef {
	return []ToolDef{
		{
			Name:        "search_commands",
			Description: h.searchDescription(),
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"query": map[string]any{
						"type":        "string",
						"description": "Search keyword (e.g., 'transaction', 'budget', 'create', 'delete')",
					},
					"resource": map[string]any{
						"type":        "string",
						"description": "Filter by resource type (e.g., 'transactions', 'accounts', 'categories', 'plans')",
					},
				},
			},
		},
		{
			Name: "run_command",
			Description: `Execute a YNAB CLI command and return JSON results. Use search_commands first to discover available commands and their flags.
Example commands: "plans list", "transactions list <plan-id> --since-date 2024-01-01", "accounts list <plan-id>".`,
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"command": map[string]any{
						"type":        "string",
						"description": "The command to execute (everything after 'ynab'), e.g. 'plans list' or 'transactions list abc-123 --since-date 2024-01-01'",
					},
				},
				"required": []string{"command"},
			},
		},
	}
}

// Call dispatches a tool call by name.
func (h *ToolHandlers) Call(name string, args json.RawMessage) (*ToolCallResult, error) {
	switch name {
	case "search_commands":
		return h.handleSearch(args)
	case "run_command":
		return h.handleRun(args)
	default:
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
}

func (h *ToolHandlers) handleSearch(args json.RawMessage) (*ToolCallResult, error) {
	var params struct {
		Query    string `json:"query"`
		Resource string `json:"resource"`
	}
	if len(args) > 0 {
		if err := json.Unmarshal(args, &params); err != nil {
			return nil, fmt.Errorf("invalid arguments: %w", err)
		}
	}

	results := h.registry.Search(params.Query, params.Resource)

	data, err := json.Marshal(results)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal results: %w", err)
	}

	return &ToolCallResult{
		Content: []ToolResult{{Type: "text", Text: string(data)}},
	}, nil
}

func (h *ToolHandlers) handleRun(args json.RawMessage) (*ToolCallResult, error) {
	var params struct {
		Command string `json:"command"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("invalid arguments: %w", err)
	}
	if params.Command == "" {
		return nil, fmt.Errorf("command is required")
	}

	cmdArgs := tokenize(params.Command)

	// Force JSON output
	cmdArgs = append(cmdArgs, "--output", "json")

	output, err := h.runner(cmdArgs)
	if err != nil {
		errMsg := err.Error()
		if output != "" {
			errMsg = output
		}
		return &ToolCallResult{
			Content: []ToolResult{{Type: "text", Text: errMsg}},
			IsError: true,
		}, nil
	}

	if output == "" {
		output = "{}"
	}

	return &ToolCallResult{
		Content: []ToolResult{{Type: "text", Text: output}},
	}, nil
}

func (h *ToolHandlers) searchDescription() string {
	resources := h.registry.Resources()
	return fmt.Sprintf(
		"Search for available YNAB budget commands. YNAB is a personal finance/budgeting tool.\n"+
			"Available resources: %s.\n"+
			"Use this to discover command names, required arguments, and flags before calling run_command.",
		strings.Join(resources, ", "),
	)
}

// tokenize splits a command string into tokens, respecting double quotes.
func tokenize(s string) []string {
	var tokens []string
	var current strings.Builder
	inQuotes := false

	for i := 0; i < len(s); i++ {
		ch := s[i]
		switch {
		case ch == '"':
			inQuotes = !inQuotes
		case ch == ' ' && !inQuotes:
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
		default:
			current.WriteByte(ch)
		}
	}
	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}
	return tokens
}
