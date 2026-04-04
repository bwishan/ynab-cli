package cli

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/bwishan/ynab-cli/internal/config"
)

func (a *App) registerCategoryCommands() {
	// ── categories command ──────────────────────────────────────────────
	catCmd := &Command{
		Name:        "categories",
		Description: "Manage budget categories",
		Subcommands: map[string]*Subcommand{},
	}

	catCmd.Subcommands["list"] = &Subcommand{
		Name:        "list",
		Description: "List all categories for a plan",
		Usage:       "<plan-id> [--last-knowledge N]",
		Run: func(ctx *Context, args []string) error {
			var planExplicit string
			var lastKnowledge int64

			positional := []string{}
			for i := 0; i < len(args); i++ {
				switch {
				case args[i] == "--last-knowledge" && i+1 < len(args):
					v, err := strconv.ParseInt(args[i+1], 10, 64)
					if err != nil {
						return fmt.Errorf("invalid --last-knowledge value: %s", args[i+1])
					}
					lastKnowledge = v
					i++
				case strings.HasPrefix(args[i], "--last-knowledge="):
					v, err := strconv.ParseInt(strings.TrimPrefix(args[i], "--last-knowledge="), 10, 64)
					if err != nil {
						return fmt.Errorf("invalid --last-knowledge value: %s", args[i])
					}
					lastKnowledge = v
				default:
					positional = append(positional, args[i])
				}
			}

			if len(positional) > 0 {
				planExplicit = positional[0]
			}

			planID, err := config.ResolvePlan(planExplicit)
			if err != nil {
				return err
			}

			data, err := ctx.Client.GetCategories(planID, lastKnowledge)
			if err != nil {
				return err
			}

			headers := []string{"ID", "Group", "Name", "Budgeted", "Activity", "Balance", "Hidden"}
			return ctx.Printer.PrintResult(data, headers, extractCategoryRows)
		},
	}

	catCmd.Subcommands["get"] = &Subcommand{
		Name:        "get",
		Description: "Get a single category",
		Usage:       "<plan-id> <category-id>",
		Run: func(ctx *Context, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("usage: ynab categories get <plan-id> <category-id>")
			}

			planID, err := config.ResolvePlan(args[0])
			if err != nil {
				return err
			}

			data, err := ctx.Client.GetCategoryByID(planID, args[1])
			if err != nil {
				return err
			}

			return ctx.Printer.PrintResult(data, nil, nil)
		},
	}

	catCmd.Subcommands["create"] = &Subcommand{
		Name:        "create",
		Description: "Create a new category",
		Usage:       "<plan-id> --name NAME --category-group-id ID",
		Run: func(ctx *Context, args []string) error {
			var planExplicit string
			var name, groupID string

			positional := []string{}
			for i := 0; i < len(args); i++ {
				switch {
				case args[i] == "--name" && i+1 < len(args):
					name = args[i+1]
					i++
				case strings.HasPrefix(args[i], "--name="):
					name = strings.TrimPrefix(args[i], "--name=")
				case args[i] == "--category-group-id" && i+1 < len(args):
					groupID = args[i+1]
					i++
				case strings.HasPrefix(args[i], "--category-group-id="):
					groupID = strings.TrimPrefix(args[i], "--category-group-id=")
				default:
					positional = append(positional, args[i])
				}
			}

			if len(positional) > 0 {
				planExplicit = positional[0]
			}

			planID, err := config.ResolvePlan(planExplicit)
			if err != nil {
				return err
			}

			if name == "" {
				return fmt.Errorf("--name is required")
			}
			if groupID == "" {
				return fmt.Errorf("--category-group-id is required")
			}

			body := map[string]interface{}{
				"category": map[string]interface{}{
					"name":              name,
					"category_group_id": groupID,
				},
			}

			data, err := ctx.Client.CreateCategory(planID, body)
			if err != nil {
				return err
			}

			return ctx.Printer.PrintResult(data, nil, nil)
		},
	}

	catCmd.Subcommands["update"] = &Subcommand{
		Name:        "update",
		Description: "Update an existing category",
		Usage:       "<plan-id> <category-id> [--name NAME] [--note NOTE]",
		Run: func(ctx *Context, args []string) error {
			var name, note string
			nameSet, noteSet := false, false

			positional := []string{}
			for i := 0; i < len(args); i++ {
				switch {
				case args[i] == "--name" && i+1 < len(args):
					name = args[i+1]
					nameSet = true
					i++
				case strings.HasPrefix(args[i], "--name="):
					name = strings.TrimPrefix(args[i], "--name=")
					nameSet = true
				case args[i] == "--note" && i+1 < len(args):
					note = args[i+1]
					noteSet = true
					i++
				case strings.HasPrefix(args[i], "--note="):
					note = strings.TrimPrefix(args[i], "--note=")
					noteSet = true
				default:
					positional = append(positional, args[i])
				}
			}

			if len(positional) < 2 {
				return fmt.Errorf("usage: ynab categories update <plan-id> <category-id> [--name NAME] [--note NOTE]")
			}

			planID, err := config.ResolvePlan(positional[0])
			if err != nil {
				return err
			}
			categoryID := positional[1]

			fields := map[string]interface{}{}
			if nameSet {
				fields["name"] = name
			}
			if noteSet {
				fields["note"] = note
			}
			if len(fields) == 0 {
				return fmt.Errorf("at least one of --name or --note is required")
			}

			body := map[string]interface{}{
				"category": fields,
			}

			data, err := ctx.Client.UpdateCategory(planID, categoryID, body)
			if err != nil {
				return err
			}

			return ctx.Printer.PrintResult(data, nil, nil)
		},
	}

	catCmd.Subcommands["get-month"] = &Subcommand{
		Name:        "get-month",
		Description: "Get a category for a specific month",
		Usage:       "<plan-id> <month> <category-id>",
		Run: func(ctx *Context, args []string) error {
			if len(args) < 3 {
				return fmt.Errorf("usage: ynab categories get-month <plan-id> <month> <category-id>")
			}

			planID, err := config.ResolvePlan(args[0])
			if err != nil {
				return err
			}

			data, err := ctx.Client.GetMonthCategory(planID, args[1], args[2])
			if err != nil {
				return err
			}

			return ctx.Printer.PrintResult(data, nil, nil)
		},
	}

	catCmd.Subcommands["update-month"] = &Subcommand{
		Name:        "update-month",
		Description: "Update a category for a specific month",
		Usage:       "<plan-id> <month> <category-id> --budgeted AMOUNT",
		Run: func(ctx *Context, args []string) error {
			var budgeted string

			positional := []string{}
			for i := 0; i < len(args); i++ {
				switch {
				case args[i] == "--budgeted" && i+1 < len(args):
					budgeted = args[i+1]
					i++
				case strings.HasPrefix(args[i], "--budgeted="):
					budgeted = strings.TrimPrefix(args[i], "--budgeted=")
				default:
					positional = append(positional, args[i])
				}
			}

			if len(positional) < 3 {
				return fmt.Errorf("usage: ynab categories update-month <plan-id> <month> <category-id> --budgeted AMOUNT")
			}

			planID, err := config.ResolvePlan(positional[0])
			if err != nil {
				return err
			}
			month := positional[1]
			categoryID := positional[2]

			if budgeted == "" {
				return fmt.Errorf("--budgeted is required")
			}

			budgetedAmount, err := strconv.ParseInt(budgeted, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid --budgeted value: %s", budgeted)
			}

			body := map[string]interface{}{
				"category": map[string]interface{}{
					"budgeted": budgetedAmount,
				},
			}

			data, err := ctx.Client.UpdateMonthCategory(planID, month, categoryID, body)
			if err != nil {
				return err
			}

			return ctx.Printer.PrintResult(data, nil, nil)
		},
	}

	a.registerCommand(catCmd)

	// ── category-groups command ─────────────────────────────────────────
	grpCmd := &Command{
		Name:        "category-groups",
		Description: "Manage category groups",
		Subcommands: map[string]*Subcommand{},
	}

	grpCmd.Subcommands["create"] = &Subcommand{
		Name:        "create",
		Description: "Create a new category group",
		Usage:       "<plan-id> --name NAME",
		Run: func(ctx *Context, args []string) error {
			var planExplicit string
			var name string

			positional := []string{}
			for i := 0; i < len(args); i++ {
				switch {
				case args[i] == "--name" && i+1 < len(args):
					name = args[i+1]
					i++
				case strings.HasPrefix(args[i], "--name="):
					name = strings.TrimPrefix(args[i], "--name=")
				default:
					positional = append(positional, args[i])
				}
			}

			if len(positional) > 0 {
				planExplicit = positional[0]
			}

			planID, err := config.ResolvePlan(planExplicit)
			if err != nil {
				return err
			}

			if name == "" {
				return fmt.Errorf("--name is required")
			}

			body := map[string]interface{}{
				"category_group": map[string]interface{}{
					"name": name,
				},
			}

			data, err := ctx.Client.CreateCategoryGroup(planID, body)
			if err != nil {
				return err
			}

			return ctx.Printer.PrintResult(data, nil, nil)
		},
	}

	grpCmd.Subcommands["update"] = &Subcommand{
		Name:        "update",
		Description: "Update a category group",
		Usage:       "<plan-id> <group-id> [--name NAME]",
		Run: func(ctx *Context, args []string) error {
			var name string
			nameSet := false

			positional := []string{}
			for i := 0; i < len(args); i++ {
				switch {
				case args[i] == "--name" && i+1 < len(args):
					name = args[i+1]
					nameSet = true
					i++
				case strings.HasPrefix(args[i], "--name="):
					name = strings.TrimPrefix(args[i], "--name=")
					nameSet = true
				default:
					positional = append(positional, args[i])
				}
			}

			if len(positional) < 2 {
				return fmt.Errorf("usage: ynab category-groups update <plan-id> <group-id> [--name NAME]")
			}

			planID, err := config.ResolvePlan(positional[0])
			if err != nil {
				return err
			}
			groupID := positional[1]

			if !nameSet {
				return fmt.Errorf("--name is required")
			}

			body := map[string]interface{}{
				"category_group": map[string]interface{}{
					"name": name,
				},
			}

			data, err := ctx.Client.UpdateCategoryGroup(planID, groupID, body)
			if err != nil {
				return err
			}

			return ctx.Printer.PrintResult(data, nil, nil)
		},
	}

	a.registerCommand(grpCmd)
}

// extractCategoryRows parses the YNAB categories response and returns rows
// for table/CSV output. The response contains category_groups, each with
// a nested categories array.
func extractCategoryRows(data []byte) ([][]string, error) {
	var resp struct {
		Data struct {
			CategoryGroups []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
				Categories []struct {
					ID       string `json:"id"`
					Name     string `json:"name"`
					Budgeted int64  `json:"budgeted"`
					Activity int64  `json:"activity"`
					Balance  int64  `json:"balance"`
					Hidden   bool   `json:"hidden"`
				} `json:"categories"`
			} `json:"category_groups"`
		} `json:"data"`
	}

	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing categories response: %w", err)
	}

	var rows [][]string
	for _, group := range resp.Data.CategoryGroups {
		for _, cat := range group.Categories {
			rows = append(rows, []string{
				cat.ID,
				group.Name,
				cat.Name,
				milliunitToString(cat.Budgeted),
				milliunitToString(cat.Activity),
				milliunitToString(cat.Balance),
				fmt.Sprintf("%t", cat.Hidden),
			})
		}
	}
	return rows, nil
}

// milliunitToString converts a YNAB milliunits value (e.g., 1234560) to a
// human-readable string (e.g., "1234.56"). YNAB stores amounts as
// milliunits (amount * 1000).
func milliunitToString(milliunits int64) string {
	whole := milliunits / 1000
	frac := milliunits % 1000
	if frac < 0 {
		frac = -frac
	}
	return fmt.Sprintf("%d.%02d", whole, frac/10)
}
