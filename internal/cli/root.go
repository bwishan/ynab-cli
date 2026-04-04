// Package cli implements the command-line interface for ynab-cli.
// It uses only the Go standard library - no external CLI frameworks.
package cli

import (
	"fmt"
	"strings"

	"github.com/bwishan/ynab-cli/internal/api"
	"github.com/bwishan/ynab-cli/internal/config"
	"github.com/bwishan/ynab-cli/internal/output"
)

// App is the root CLI application.
type App struct {
	Version   string
	Commit    string
	BuildDate string
	commands  map[string]*Command
}

// Command represents a CLI command with subcommands.
type Command struct {
	Name        string
	Description string
	Subcommands map[string]*Subcommand
}

// Subcommand represents a leaf action (e.g., "list", "get", "create").
type Subcommand struct {
	Name        string
	Description string
	Usage       string
	Run         func(ctx *Context, args []string) error
}

// Context holds resolved runtime state passed to command handlers.
type Context struct {
	Client  *api.Client
	Printer *output.Printer
	Token   string
}

// New creates the root CLI app and registers all commands.
func New(version, commit, buildDate string) *App {
	app := &App{
		Version:   version,
		Commit:    commit,
		BuildDate: buildDate,
		commands:  make(map[string]*Command),
	}

	app.registerUserCommands()
	app.registerPlanCommands()
	app.registerAccountCommands()
	app.registerCategoryCommands()
	app.registerTransactionCommands()
	app.registerPayeeCommands()
	app.registerScheduledCommands()
	app.registerMonthCommands()
	app.registerMoneyMovementCommands()

	return app
}

// Run parses the CLI arguments and dispatches to the appropriate handler.
func (a *App) Run(args []string) error {
	// Parse global flags
	token := ""
	formatStr := "json"
	pretty := false
	var remaining []string

	for i := 0; i < len(args); i++ {
		switch {
		case args[i] == "--token" && i+1 < len(args):
			token = args[i+1]
			i++
		case strings.HasPrefix(args[i], "--token="):
			token = strings.TrimPrefix(args[i], "--token=")
		case args[i] == "--output" && i+1 < len(args):
			formatStr = args[i+1]
			i++
		case strings.HasPrefix(args[i], "--output="):
			formatStr = strings.TrimPrefix(args[i], "--output=")
		case args[i] == "--pretty":
			pretty = true
		case args[i] == "--version" || args[i] == "-v":
			fmt.Printf("ynab %s (commit: %s, built: %s)\n", a.Version, a.Commit, a.BuildDate)
			return nil
		case args[i] == "configure":
			return a.runConfigure(args[i+1:])
		case (args[i] == "--help" || args[i] == "-h" || args[i] == "help") && len(remaining) == 0:
			a.printHelp()
			return nil
		default:
			// Once we hit a non-global-flag, collect it and all remaining args
			remaining = append(remaining, args[i:]...)
			i = len(args)
		}
	}

	if len(remaining) == 0 {
		a.printHelp()
		return nil
	}

	format := output.Format(formatStr)
	switch format {
	case output.FormatJSON, output.FormatTable, output.FormatCSV:
	default:
		return fmt.Errorf("invalid output format %q (valid: json, table, csv)", formatStr)
	}

	cmdName := remaining[0]
	cmd, ok := a.commands[cmdName]
	if !ok {
		return fmt.Errorf("unknown command %q. Run 'ynab --help' for usage", cmdName)
	}

	if len(remaining) < 2 || remaining[1] == "--help" || remaining[1] == "-h" {
		a.printCommandHelp(cmd)
		return nil
	}

	subName := remaining[1]
	sub, ok := cmd.Subcommands[subName]
	if !ok {
		return fmt.Errorf("unknown subcommand %q for %q. Run 'ynab %s --help' for usage", subName, cmdName, cmdName)
	}

	resolvedToken, err := config.ResolveToken(token)
	if err != nil {
		return err
	}

	ctx := &Context{
		Client:  api.NewClient(resolvedToken),
		Printer: output.NewPrinter(format, pretty),
		Token:   resolvedToken,
	}

	return sub.Run(ctx, remaining[2:])
}

// registerCommand adds a command group to the app.
func (a *App) registerCommand(cmd *Command) {
	a.commands[cmd.Name] = cmd
}

func (a *App) printHelp() {
	fmt.Println(`ynab - Agent-friendly CLI for the YNAB API

USAGE:
  ynab [global-flags] <command> <subcommand> [arguments]

GLOBAL FLAGS:
  --token <token>      YNAB access token (or set YNAB_ACCESS_TOKEN)
  --output <format>    Output format: json (default), table, csv
  --pretty             Pretty-print JSON output
  --version, -v        Print version
  --help, -h           Print help

COMMANDS:`)

	order := []string{"user", "plans", "accounts", "categories", "category-groups",
		"transactions", "payees", "payee-locations", "scheduled", "months",
		"money-movements", "money-movement-groups"}
	for _, name := range order {
		if cmd, ok := a.commands[name]; ok {
			fmt.Printf("  %-24s %s\n", cmd.Name, cmd.Description)
		}
	}

	fmt.Println(`
CONFIGURATION:
  ynab configure --token <token>              Save access token
  ynab configure --default-plan <plan-id>     Set default plan

ENVIRONMENT VARIABLES:
  YNAB_ACCESS_TOKEN    Access token for authentication
  YNAB_PLAN_ID         Default plan ID

EXAMPLES:
  ynab plans list
  ynab accounts list <plan-id>
  ynab transactions list <plan-id> --since-date 2024-01-01
  ynab categories list <plan-id> --output table`)
}

func (a *App) printCommandHelp(cmd *Command) {
	fmt.Printf("ynab %s - %s\n\nSUBCOMMANDS:\n", cmd.Name, cmd.Description)
	for _, sub := range cmd.Subcommands {
		fmt.Printf("  %-20s %s\n", sub.Name, sub.Description)
		if sub.Usage != "" {
			fmt.Printf("  %-20s Usage: ynab %s %s %s\n", "", cmd.Name, sub.Name, sub.Usage)
		}
	}
}

func (a *App) runConfigure(args []string) error {
	cfg, err := config.Load()
	if err != nil {
		cfg = &config.Config{}
	}

	for i := 0; i < len(args); i++ {
		switch {
		case args[i] == "--token" && i+1 < len(args):
			cfg.AccessToken = args[i+1]
			i++
		case strings.HasPrefix(args[i], "--token="):
			cfg.AccessToken = strings.TrimPrefix(args[i], "--token=")
		case args[i] == "--default-plan" && i+1 < len(args):
			cfg.DefaultPlan = args[i+1]
			i++
		case strings.HasPrefix(args[i], "--default-plan="):
			cfg.DefaultPlan = strings.TrimPrefix(args[i], "--default-plan=")
		default:
			return fmt.Errorf("unknown configure flag: %s", args[i])
		}
	}

	if err := config.Save(cfg); err != nil {
		return err
	}
	fmt.Println("Configuration saved.")
	return nil
}
