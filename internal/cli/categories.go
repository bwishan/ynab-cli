package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/bwishan/ynab-cli/internal/config"
)

func (a *App) registerCategoryCommands() {
	// ── categories command ──────────────────────────────────────────────
	catCmd := &cobra.Command{
		Use:   "categories",
		Short: "Manage budget categories",
	}

	listCmd := &cobra.Command{
		Use:   "list [plan-id]",
		Short: "List all categories for a plan",
		RunE: func(cmd *cobra.Command, args []string) error {
			var planIDArg string
			if len(args) > 0 {
				planIDArg = args[0]
			}

			planID, err := config.ResolvePlan(planIDArg)
			if err != nil {
				return err
			}

			lastKnowledge, _ := cmd.Flags().GetInt64("last-knowledge")

			data, err := a.client.GetCategories(planID, lastKnowledge)
			if err != nil {
				return err
			}

			headers := []string{"ID", "Group", "Name", "Budgeted", "Activity", "Balance", "Hidden"}
			return a.printer.PrintResult(data, headers, extractCategoryRows)
		},
	}
	listCmd.Flags().Int64("last-knowledge", 0, "Server knowledge for delta sync")

	getCmd := &cobra.Command{
		Use:   "get [plan-id] <category-id>",
		Short: "Get a single category",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("usage: ynab categories get <plan-id> <category-id>")
			}

			planID, err := config.ResolvePlan(args[0])
			if err != nil {
				return err
			}

			data, err := a.client.GetCategoryByID(planID, args[1])
			if err != nil {
				return err
			}

			return a.printer.PrintResult(data, nil, nil)
		},
	}

	createCmd := &cobra.Command{
		Use:   "create [plan-id]",
		Short: "Create a new category",
		RunE: func(cmd *cobra.Command, args []string) error {
			var planIDArg string
			if len(args) > 0 {
				planIDArg = args[0]
			}

			planID, err := config.ResolvePlan(planIDArg)
			if err != nil {
				return err
			}

			name, _ := cmd.Flags().GetString("name")
			if name == "" {
				return fmt.Errorf("--name is required")
			}

			groupID, _ := cmd.Flags().GetString("category-group-id")
			if groupID == "" {
				return fmt.Errorf("--category-group-id is required")
			}

			body := map[string]interface{}{
				"category": map[string]interface{}{
					"name":              name,
					"category_group_id": groupID,
				},
			}

			data, err := a.client.CreateCategory(planID, body)
			if err != nil {
				return err
			}

			return a.printer.PrintResult(data, nil, nil)
		},
	}
	createCmd.Flags().String("name", "", "Category name")
	createCmd.Flags().String("category-group-id", "", "Category group ID")

	updateCmd := &cobra.Command{
		Use:   "update [plan-id] <category-id>",
		Short: "Update an existing category",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("usage: ynab categories update <plan-id> <category-id> [--name NAME] [--note NOTE]")
			}

			planID, err := config.ResolvePlan(args[0])
			if err != nil {
				return err
			}
			categoryID := args[1]

			fields := map[string]interface{}{}
			if cmd.Flags().Changed("name") {
				name, _ := cmd.Flags().GetString("name")
				fields["name"] = name
			}
			if cmd.Flags().Changed("note") {
				note, _ := cmd.Flags().GetString("note")
				fields["note"] = note
			}
			if len(fields) == 0 {
				return fmt.Errorf("at least one of --name or --note is required")
			}

			body := map[string]interface{}{
				"category": fields,
			}

			data, err := a.client.UpdateCategory(planID, categoryID, body)
			if err != nil {
				return err
			}

			return a.printer.PrintResult(data, nil, nil)
		},
	}
	updateCmd.Flags().String("name", "", "Category name")
	updateCmd.Flags().String("note", "", "Category note")

	getMonthCmd := &cobra.Command{
		Use:   "get-month [plan-id] <month> <category-id>",
		Short: "Get a category for a specific month",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 3 {
				return fmt.Errorf("usage: ynab categories get-month <plan-id> <month> <category-id>")
			}

			planID, err := config.ResolvePlan(args[0])
			if err != nil {
				return err
			}

			data, err := a.client.GetMonthCategory(planID, args[1], args[2])
			if err != nil {
				return err
			}

			return a.printer.PrintResult(data, nil, nil)
		},
	}

	updateMonthCmd := &cobra.Command{
		Use:   "update-month [plan-id] <month> <category-id>",
		Short: "Update a category for a specific month",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 3 {
				return fmt.Errorf("usage: ynab categories update-month <plan-id> <month> <category-id> --budgeted AMOUNT")
			}

			planID, err := config.ResolvePlan(args[0])
			if err != nil {
				return err
			}
			month := args[1]
			categoryID := args[2]

			budgeted, _ := cmd.Flags().GetInt64("budgeted")
			if !cmd.Flags().Changed("budgeted") {
				return fmt.Errorf("--budgeted is required")
			}

			body := map[string]interface{}{
				"category": map[string]interface{}{
					"budgeted": budgeted,
				},
			}

			data, err := a.client.UpdateMonthCategory(planID, month, categoryID, body)
			if err != nil {
				return err
			}

			return a.printer.PrintResult(data, nil, nil)
		},
	}
	updateMonthCmd.Flags().Int64("budgeted", 0, "Budgeted amount in milliunits")

	catCmd.AddCommand(listCmd, getCmd, createCmd, updateCmd, getMonthCmd, updateMonthCmd)
	a.rootCmd.AddCommand(catCmd)

	// ── category-groups command ─────────────────────────────────────────
	grpCmd := &cobra.Command{
		Use:   "category-groups",
		Short: "Manage category groups",
	}

	grpCreateCmd := &cobra.Command{
		Use:   "create [plan-id]",
		Short: "Create a new category group",
		RunE: func(cmd *cobra.Command, args []string) error {
			var planIDArg string
			if len(args) > 0 {
				planIDArg = args[0]
			}

			planID, err := config.ResolvePlan(planIDArg)
			if err != nil {
				return err
			}

			name, _ := cmd.Flags().GetString("name")
			if name == "" {
				return fmt.Errorf("--name is required")
			}

			body := map[string]interface{}{
				"category_group": map[string]interface{}{
					"name": name,
				},
			}

			data, err := a.client.CreateCategoryGroup(planID, body)
			if err != nil {
				return err
			}

			return a.printer.PrintResult(data, nil, nil)
		},
	}
	grpCreateCmd.Flags().String("name", "", "Category group name")

	grpUpdateCmd := &cobra.Command{
		Use:   "update [plan-id] <group-id>",
		Short: "Update a category group",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("usage: ynab category-groups update <plan-id> <group-id> --name NAME")
			}

			planID, err := config.ResolvePlan(args[0])
			if err != nil {
				return err
			}
			groupID := args[1]

			name, _ := cmd.Flags().GetString("name")
			if name == "" {
				return fmt.Errorf("--name is required")
			}

			body := map[string]interface{}{
				"category_group": map[string]interface{}{
					"name": name,
				},
			}

			data, err := a.client.UpdateCategoryGroup(planID, groupID, body)
			if err != nil {
				return err
			}

			return a.printer.PrintResult(data, nil, nil)
		},
	}
	grpUpdateCmd.Flags().String("name", "", "Category group name")

	grpCmd.AddCommand(grpCreateCmd, grpUpdateCmd)
	a.rootCmd.AddCommand(grpCmd)
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
