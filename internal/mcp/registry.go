package mcp

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// FlagInfo describes a command flag.
type FlagInfo struct {
	Name        string `json:"name"`
	Shorthand   string `json:"shorthand,omitempty"`
	Type        string `json:"type"`
	Default     string `json:"default,omitempty"`
	Description string `json:"description"`
	Required    bool   `json:"required,omitempty"`
}

// ArgInfo describes a positional argument.
type ArgInfo struct {
	Name     string `json:"name"`
	Required bool   `json:"required"`
}

// CommandInfo describes a CLI command.
type CommandInfo struct {
	Name        string     `json:"name"`
	Usage       string     `json:"usage"`
	Description string     `json:"description"`
	Resource    string     `json:"resource"`
	Action      string     `json:"action"`
	Args        []ArgInfo  `json:"args,omitempty"`
	Flags       []FlagInfo `json:"flags,omitempty"`
}

// Registry holds a searchable index of CLI commands.
type Registry struct {
	commands []CommandInfo
}

// NewRegistry builds a command index by walking the cobra command tree.
func NewRegistry(rootCmd *cobra.Command) *Registry {
	r := &Registry{}
	r.walk(rootCmd, nil)
	return r
}

func (r *Registry) walk(cmd *cobra.Command, path []string) {
	for _, child := range cmd.Commands() {
		childPath := append(append([]string{}, path...), child.Name())

		// Skip hidden, help, and specific commands
		if child.Hidden || child.Name() == "help" || child.Name() == "completion" {
			continue
		}
		// Skip mcp and configure from the index
		if child.Name() == "mcp" || child.Name() == "configure" {
			continue
		}

		if child.RunE != nil || child.Run != nil {
			// Leaf command — index it
			info := CommandInfo{
				Name:        strings.Join(childPath, " "),
				Usage:       "ynab " + strings.Join(childPath[:len(childPath)-1], " ") + " " + child.Use,
				Description: child.Short,
			}

			if len(childPath) >= 1 {
				info.Resource = childPath[0]
			}
			if len(childPath) >= 2 {
				info.Action = childPath[1]
			}

			info.Args = parseArgs(child.Use)
			info.Flags = extractFlags(child)

			r.commands = append(r.commands, info)
		}

		// Recurse into subcommands
		r.walk(child, childPath)
	}
}

// parseArgs extracts positional arguments from the Use string.
// Patterns: <arg> is required, [arg] is optional.
func parseArgs(use string) []ArgInfo {
	var args []ArgInfo
	// Split the Use string and look for <arg> or [arg] patterns
	parts := strings.Fields(use)
	for _, p := range parts[1:] { // skip the command name itself
		if strings.HasPrefix(p, "--") {
			break // flags section
		}
		if strings.HasPrefix(p, "<") && strings.HasSuffix(p, ">") {
			args = append(args, ArgInfo{
				Name:     strings.Trim(p, "<>"),
				Required: true,
			})
		} else if strings.HasPrefix(p, "[") && strings.HasSuffix(p, "]") {
			args = append(args, ArgInfo{
				Name:     strings.Trim(p, "[]"),
				Required: false,
			})
		}
	}
	return args
}

func extractFlags(cmd *cobra.Command) []FlagInfo {
	var flags []FlagInfo
	// Collect local (non-inherited) flags only
	cmd.LocalFlags().VisitAll(func(f *pflag.Flag) {
		// Skip help flag
		if f.Name == "help" {
			return
		}
		fi := FlagInfo{
			Name:        f.Name,
			Shorthand:   f.Shorthand,
			Type:        f.Value.Type(),
			Default:     f.DefValue,
			Description: f.Usage,
		}
		// Check cobra required annotation
		if ann, ok := f.Annotations[cobra.BashCompOneRequiredFlag]; ok && len(ann) > 0 && ann[0] == "true" {
			fi.Required = true
		}
		flags = append(flags, fi)
	})
	return flags
}

// Search finds commands matching the query and/or resource filter.
func (r *Registry) Search(query, resource string) []CommandInfo {
	if query == "" && resource == "" {
		return r.commands
	}

	query = strings.ToLower(query)
	resource = strings.ToLower(resource)

	var results []CommandInfo
	for _, cmd := range r.commands {
		if resource != "" && strings.ToLower(cmd.Resource) != resource {
			continue
		}
		if query != "" && !matchesQuery(cmd, query) {
			continue
		}
		results = append(results, cmd)
	}
	return results
}

func matchesQuery(cmd CommandInfo, query string) bool {
	fields := []string{
		cmd.Name,
		cmd.Description,
		cmd.Resource,
		cmd.Action,
	}
	for _, f := range cmd.Flags {
		fields = append(fields, f.Name, f.Description)
	}
	for _, f := range fields {
		if strings.Contains(strings.ToLower(f), query) {
			return true
		}
	}
	return false
}

// Commands returns all indexed commands.
func (r *Registry) Commands() []CommandInfo {
	return r.commands
}

// Resources returns a deduplicated, sorted list of resource names.
func (r *Registry) Resources() []string {
	seen := make(map[string]bool)
	var resources []string
	for _, cmd := range r.commands {
		if cmd.Resource != "" && !seen[cmd.Resource] {
			seen[cmd.Resource] = true
			resources = append(resources, cmd.Resource)
		}
	}
	return resources
}
