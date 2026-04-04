package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/bwishan/ynab-cli/internal/config"
)

func (a *App) registerPayeeCommands() {
	// ── payees command ──────────────────────────────────────────────────
	payeesCmd := &cobra.Command{
		Use:   "payees",
		Short: "Manage payees",
	}

	payeesListCmd := &cobra.Command{
		Use:   "list [plan-id]",
		Short: "List all payees for a plan",
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

			data, err := a.client.GetPayees(planID, lastKnowledge)
			if err != nil {
				return err
			}

			headers := []string{"ID", "NAME", "TRANSFER_ACCOUNT_ID", "DELETED"}
			extractRows := func(data []byte) ([][]string, error) {
				var resp struct {
					Data struct {
						Payees []struct {
							ID                string `json:"id"`
							Name              string `json:"name"`
							TransferAccountID string `json:"transfer_account_id"`
							Deleted           bool   `json:"deleted"`
						} `json:"payees"`
					} `json:"data"`
				}
				if err := json.Unmarshal(data, &resp); err != nil {
					return nil, err
				}
				var rows [][]string
				for _, p := range resp.Data.Payees {
					rows = append(rows, []string{
						p.ID,
						p.Name,
						p.TransferAccountID,
						fmt.Sprintf("%t", p.Deleted),
					})
				}
				return rows, nil
			}

			return a.printer.PrintResult(data, headers, extractRows)
		},
	}
	payeesListCmd.Flags().Int64("last-knowledge", 0, "Server knowledge for delta sync")

	payeesGetCmd := &cobra.Command{
		Use:   "get [plan-id] <payee-id>",
		Short: "Get a specific payee",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("usage: ynab payees get <plan-id> <payee-id>")
			}

			planID, err := config.ResolvePlan(args[0])
			if err != nil {
				return err
			}

			data, err := a.client.GetPayeeByID(planID, args[1])
			if err != nil {
				return err
			}

			headers := []string{"ID", "NAME", "TRANSFER_ACCOUNT_ID", "DELETED"}
			extractRows := func(data []byte) ([][]string, error) {
				var resp struct {
					Data struct {
						Payee struct {
							ID                string `json:"id"`
							Name              string `json:"name"`
							TransferAccountID string `json:"transfer_account_id"`
							Deleted           bool   `json:"deleted"`
						} `json:"payee"`
					} `json:"data"`
				}
				if err := json.Unmarshal(data, &resp); err != nil {
					return nil, err
				}
				p := resp.Data.Payee
				return [][]string{{
					p.ID,
					p.Name,
					p.TransferAccountID,
					fmt.Sprintf("%t", p.Deleted),
				}}, nil
			}

			return a.printer.PrintResult(data, headers, extractRows)
		},
	}

	payeesCreateCmd := &cobra.Command{
		Use:   "create [plan-id]",
		Short: "Create a new payee",
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
				"payee": map[string]interface{}{
					"name": name,
				},
			}

			data, err := a.client.CreatePayee(planID, body)
			if err != nil {
				return err
			}

			headers := []string{"ID", "NAME"}
			extractRows := func(data []byte) ([][]string, error) {
				var resp struct {
					Data struct {
						Payee struct {
							ID   string `json:"id"`
							Name string `json:"name"`
						} `json:"payee"`
					} `json:"data"`
				}
				if err := json.Unmarshal(data, &resp); err != nil {
					return nil, err
				}
				p := resp.Data.Payee
				return [][]string{{p.ID, p.Name}}, nil
			}

			return a.printer.PrintResult(data, headers, extractRows)
		},
	}
	payeesCreateCmd.Flags().String("name", "", "Payee name")

	payeesUpdateCmd := &cobra.Command{
		Use:   "update [plan-id] <payee-id>",
		Short: "Update a payee",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("usage: ynab payees update <plan-id> <payee-id> --name NAME")
			}

			planID, err := config.ResolvePlan(args[0])
			if err != nil {
				return err
			}

			name, _ := cmd.Flags().GetString("name")
			if name == "" {
				return fmt.Errorf("--name is required")
			}

			body := map[string]interface{}{
				"payee": map[string]interface{}{
					"name": name,
				},
			}

			data, err := a.client.UpdatePayee(planID, args[1], body)
			if err != nil {
				return err
			}

			headers := []string{"ID", "NAME"}
			extractRows := func(data []byte) ([][]string, error) {
				var resp struct {
					Data struct {
						Payee struct {
							ID   string `json:"id"`
							Name string `json:"name"`
						} `json:"payee"`
					} `json:"data"`
				}
				if err := json.Unmarshal(data, &resp); err != nil {
					return nil, err
				}
				p := resp.Data.Payee
				return [][]string{{p.ID, p.Name}}, nil
			}

			return a.printer.PrintResult(data, headers, extractRows)
		},
	}
	payeesUpdateCmd.Flags().String("name", "", "Payee name")

	payeesCmd.AddCommand(payeesListCmd, payeesGetCmd, payeesCreateCmd, payeesUpdateCmd)
	a.rootCmd.AddCommand(payeesCmd)

	// ── payee-locations command ─────────────────────────────────────────
	locationsCmd := &cobra.Command{
		Use:   "payee-locations",
		Short: "Manage payee locations",
	}

	locListCmd := &cobra.Command{
		Use:   "list [plan-id]",
		Short: "List all payee locations for a plan",
		RunE: func(cmd *cobra.Command, args []string) error {
			var planIDArg string
			if len(args) > 0 {
				planIDArg = args[0]
			}

			planID, err := config.ResolvePlan(planIDArg)
			if err != nil {
				return err
			}

			data, err := a.client.GetPayeeLocations(planID)
			if err != nil {
				return err
			}

			headers := []string{"ID", "PAYEE_ID", "LATITUDE", "LONGITUDE", "DELETED"}
			extractRows := func(data []byte) ([][]string, error) {
				var resp struct {
					Data struct {
						PayeeLocations []struct {
							ID        string `json:"id"`
							PayeeID   string `json:"payee_id"`
							Latitude  string `json:"latitude"`
							Longitude string `json:"longitude"`
							Deleted   bool   `json:"deleted"`
						} `json:"payee_locations"`
					} `json:"data"`
				}
				if err := json.Unmarshal(data, &resp); err != nil {
					return nil, err
				}
				var rows [][]string
				for _, l := range resp.Data.PayeeLocations {
					rows = append(rows, []string{
						l.ID,
						l.PayeeID,
						l.Latitude,
						l.Longitude,
						fmt.Sprintf("%t", l.Deleted),
					})
				}
				return rows, nil
			}

			return a.printer.PrintResult(data, headers, extractRows)
		},
	}

	locGetCmd := &cobra.Command{
		Use:   "get [plan-id] <location-id>",
		Short: "Get a specific payee location",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("usage: ynab payee-locations get <plan-id> <location-id>")
			}

			planID, err := config.ResolvePlan(args[0])
			if err != nil {
				return err
			}

			data, err := a.client.GetPayeeLocationByID(planID, args[1])
			if err != nil {
				return err
			}

			headers := []string{"ID", "PAYEE_ID", "LATITUDE", "LONGITUDE", "DELETED"}
			extractRows := func(data []byte) ([][]string, error) {
				var resp struct {
					Data struct {
						PayeeLocation struct {
							ID        string `json:"id"`
							PayeeID   string `json:"payee_id"`
							Latitude  string `json:"latitude"`
							Longitude string `json:"longitude"`
							Deleted   bool   `json:"deleted"`
						} `json:"payee_location"`
					} `json:"data"`
				}
				if err := json.Unmarshal(data, &resp); err != nil {
					return nil, err
				}
				l := resp.Data.PayeeLocation
				return [][]string{{
					l.ID,
					l.PayeeID,
					l.Latitude,
					l.Longitude,
					fmt.Sprintf("%t", l.Deleted),
				}}, nil
			}

			return a.printer.PrintResult(data, headers, extractRows)
		},
	}

	locListByPayeeCmd := &cobra.Command{
		Use:   "list-by-payee [plan-id] <payee-id>",
		Short: "List locations for a specific payee",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("usage: ynab payee-locations list-by-payee <plan-id> <payee-id>")
			}

			planID, err := config.ResolvePlan(args[0])
			if err != nil {
				return err
			}

			data, err := a.client.GetPayeeLocationsByPayee(planID, args[1])
			if err != nil {
				return err
			}

			headers := []string{"ID", "PAYEE_ID", "LATITUDE", "LONGITUDE", "DELETED"}
			extractRows := func(data []byte) ([][]string, error) {
				var resp struct {
					Data struct {
						PayeeLocations []struct {
							ID        string `json:"id"`
							PayeeID   string `json:"payee_id"`
							Latitude  string `json:"latitude"`
							Longitude string `json:"longitude"`
							Deleted   bool   `json:"deleted"`
						} `json:"payee_locations"`
					} `json:"data"`
				}
				if err := json.Unmarshal(data, &resp); err != nil {
					return nil, err
				}
				var rows [][]string
				for _, l := range resp.Data.PayeeLocations {
					rows = append(rows, []string{
						l.ID,
						l.PayeeID,
						l.Latitude,
						l.Longitude,
						fmt.Sprintf("%t", l.Deleted),
					})
				}
				return rows, nil
			}

			return a.printer.PrintResult(data, headers, extractRows)
		},
	}

	locationsCmd.AddCommand(locListCmd, locGetCmd, locListByPayeeCmd)
	a.rootCmd.AddCommand(locationsCmd)
}
