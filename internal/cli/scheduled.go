package cli

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/bwishan/ynab-cli/internal/config"
)

func (a *App) registerScheduledCommands() {
	schedCmd := &cobra.Command{
		Use:   "scheduled",
		Short: "Manage scheduled transactions",
	}

	// ── list ────────────────────────────────────────────────────────────
	listCmd := &cobra.Command{
		Use:   "list [plan-id]",
		Short: "List all scheduled transactions for a plan",
		Args:  cobra.MaximumNArgs(1),
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

			data, err := a.client.GetScheduledTransactions(planID, lastKnowledge)
			if err != nil {
				return err
			}

			headers := []string{"ID", "DATE_FIRST", "DATE_NEXT", "FREQUENCY", "AMOUNT", "ACCOUNT_ID", "PAYEE_ID", "CATEGORY_ID", "MEMO", "FLAG_COLOR", "DELETED"}
			extractRows := func(data []byte) ([][]string, error) {
				var resp struct {
					Data struct {
						ScheduledTransactions []struct {
							ID         string `json:"id"`
							DateFirst  string `json:"date_first"`
							DateNext   string `json:"date_next"`
							Frequency  string `json:"frequency"`
							Amount     int64  `json:"amount"`
							AccountID  string `json:"account_id"`
							PayeeID    string `json:"payee_id"`
							CategoryID string `json:"category_id"`
							Memo       string `json:"memo"`
							FlagColor  string `json:"flag_color"`
							Deleted    bool   `json:"deleted"`
						} `json:"scheduled_transactions"`
					} `json:"data"`
				}
				if err := json.Unmarshal(data, &resp); err != nil {
					return nil, err
				}
				var rows [][]string
				for _, t := range resp.Data.ScheduledTransactions {
					rows = append(rows, []string{
						t.ID,
						t.DateFirst,
						t.DateNext,
						t.Frequency,
						strconv.FormatInt(t.Amount, 10),
						t.AccountID,
						t.PayeeID,
						t.CategoryID,
						t.Memo,
						t.FlagColor,
						fmt.Sprintf("%t", t.Deleted),
					})
				}
				return rows, nil
			}

			return a.printer.PrintResult(data, headers, extractRows)
		},
	}
	listCmd.Flags().Int64("last-knowledge", 0, "Server knowledge for delta sync")

	// ── get ─────────────────────────────────────────────────────────────
	getCmd := &cobra.Command{
		Use:   "get <plan-id> <id>",
		Short: "Get a specific scheduled transaction",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			planID, err := config.ResolvePlan(args[0])
			if err != nil {
				return err
			}

			data, err := a.client.GetScheduledTransactionByID(planID, args[1])
			if err != nil {
				return err
			}

			headers := []string{"ID", "DATE_FIRST", "DATE_NEXT", "FREQUENCY", "AMOUNT", "ACCOUNT_ID", "PAYEE_ID", "CATEGORY_ID", "MEMO", "FLAG_COLOR", "DELETED"}
			extractRows := func(data []byte) ([][]string, error) {
				var resp struct {
					Data struct {
						ScheduledTransaction struct {
							ID         string `json:"id"`
							DateFirst  string `json:"date_first"`
							DateNext   string `json:"date_next"`
							Frequency  string `json:"frequency"`
							Amount     int64  `json:"amount"`
							AccountID  string `json:"account_id"`
							PayeeID    string `json:"payee_id"`
							CategoryID string `json:"category_id"`
							Memo       string `json:"memo"`
							FlagColor  string `json:"flag_color"`
							Deleted    bool   `json:"deleted"`
						} `json:"scheduled_transaction"`
					} `json:"data"`
				}
				if err := json.Unmarshal(data, &resp); err != nil {
					return nil, err
				}
				t := resp.Data.ScheduledTransaction
				return [][]string{{
					t.ID,
					t.DateFirst,
					t.DateNext,
					t.Frequency,
					strconv.FormatInt(t.Amount, 10),
					t.AccountID,
					t.PayeeID,
					t.CategoryID,
					t.Memo,
					t.FlagColor,
					fmt.Sprintf("%t", t.Deleted),
				}}, nil
			}

			return a.printer.PrintResult(data, headers, extractRows)
		},
	}

	// ── create ──────────────────────────────────────────────────────────
	createCmd := &cobra.Command{
		Use:   "create [plan-id]",
		Short: "Create a scheduled transaction",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var planIDArg string
			if len(args) > 0 {
				planIDArg = args[0]
			}

			frequency, _ := cmd.Flags().GetString("frequency")
			if err := validateFrequency(frequency); err != nil {
				return err
			}

			planID, err := config.ResolvePlan(planIDArg)
			if err != nil {
				return err
			}

			accountID, _ := cmd.Flags().GetString("account-id")
			date, _ := cmd.Flags().GetString("date")
			amount, _ := cmd.Flags().GetInt64("amount")
			payeeID, _ := cmd.Flags().GetString("payee-id")
			categoryID, _ := cmd.Flags().GetString("category-id")
			memo, _ := cmd.Flags().GetString("memo")
			flagColor, _ := cmd.Flags().GetString("flag-color")

			fields := map[string]interface{}{
				"account_id": accountID,
				"date": date,
				"amount":     amount,
				"frequency":  frequency,
			}
			if payeeID != "" {
				fields["payee_id"] = payeeID
			}
			if categoryID != "" {
				fields["category_id"] = categoryID
			}
			if memo != "" {
				fields["memo"] = memo
			}
			if flagColor != "" {
				fields["flag_color"] = flagColor
			}

			body := map[string]interface{}{
				"scheduled_transaction": fields,
			}

			data, err := a.client.CreateScheduledTransaction(planID, body)
			if err != nil {
				return err
			}
			return a.printer.PrintResult(data, nil, nil)
		},
	}
	createCmd.Flags().String("account-id", "", "Account ID (required)")
	createCmd.Flags().String("date", "", "First occurrence date (required)")
	createCmd.Flags().Int64("amount", 0, "Amount in milliunits, negative for outflows (required)")
	createCmd.Flags().String("frequency", "", "Recurrence frequency: never, daily, weekly, everyOtherWeek, twiceAMonth, every4Weeks, monthly, everyOtherMonth, every3Months, every4Months, twiceAYear, yearly, everyOtherYear (required)")
	createCmd.Flags().String("payee-id", "", "Payee ID")
	createCmd.Flags().String("category-id", "", "Category ID")
	createCmd.Flags().String("memo", "", "Memo")
	createCmd.Flags().String("flag-color", "", "Flag color")
	_ = createCmd.MarkFlagRequired("account-id")
	_ = createCmd.MarkFlagRequired("date")
	_ = createCmd.MarkFlagRequired("amount")
	_ = createCmd.MarkFlagRequired("frequency")

	// ── update ──────────────────────────────────────────────────────────
	updateCmd := &cobra.Command{
		Use:   "update <plan-id> <id>",
		Short: "Update a scheduled transaction",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			planID, err := config.ResolvePlan(args[0])
			if err != nil {
				return err
			}
			transactionID := args[1]

			frequency, _ := cmd.Flags().GetString("frequency")
			if cmd.Flags().Changed("frequency") {
				if err := validateFrequency(frequency); err != nil {
					return err
				}
			}

			accountID, _ := cmd.Flags().GetString("account-id")
			date, _ := cmd.Flags().GetString("date")
			amount, _ := cmd.Flags().GetInt64("amount")
			payeeID, _ := cmd.Flags().GetString("payee-id")
			categoryID, _ := cmd.Flags().GetString("category-id")
			memo, _ := cmd.Flags().GetString("memo")
			flagColor, _ := cmd.Flags().GetString("flag-color")

			fields := map[string]interface{}{}
			if cmd.Flags().Changed("account-id") {
				fields["account_id"] = accountID
			}
			if cmd.Flags().Changed("date") {
				fields["date"] = date
			}
			if cmd.Flags().Changed("amount") {
				fields["amount"] = amount
			}
			if cmd.Flags().Changed("frequency") {
				fields["frequency"] = frequency
			}
			if cmd.Flags().Changed("payee-id") {
				fields["payee_id"] = payeeID
			}
			if cmd.Flags().Changed("category-id") {
				fields["category_id"] = categoryID
			}
			if cmd.Flags().Changed("memo") {
				fields["memo"] = memo
			}
			if cmd.Flags().Changed("flag-color") {
				fields["flag_color"] = flagColor
			}

			if len(fields) == 0 {
				return fmt.Errorf("at least one flag is required for update")
			}

			body := map[string]interface{}{
				"scheduled_transaction": fields,
			}

			data, err := a.client.UpdateScheduledTransaction(planID, transactionID, body)
			if err != nil {
				return err
			}
			return a.printer.PrintResult(data, nil, nil)
		},
	}
	updateCmd.Flags().String("account-id", "", "Account ID")
	updateCmd.Flags().String("date", "", "First occurrence date")
	updateCmd.Flags().Int64("amount", 0, "Amount in milliunits, negative for outflows")
	updateCmd.Flags().String("frequency", "", "Recurrence frequency: never, daily, weekly, everyOtherWeek, twiceAMonth, every4Weeks, monthly, everyOtherMonth, every3Months, every4Months, twiceAYear, yearly, everyOtherYear")
	updateCmd.Flags().String("payee-id", "", "Payee ID")
	updateCmd.Flags().String("category-id", "", "Category ID")
	updateCmd.Flags().String("memo", "", "Memo")
	updateCmd.Flags().String("flag-color", "", "Flag color")

	// ── delete ──────────────────────────────────────────────────────────
	deleteCmd := &cobra.Command{
		Use:   "delete <plan-id> <id>",
		Short: "Delete a scheduled transaction",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			planID, err := config.ResolvePlan(args[0])
			if err != nil {
				return err
			}
			data, err := a.client.DeleteScheduledTransaction(planID, args[1])
			if err != nil {
				return err
			}
			return a.printer.PrintResult(data, nil, nil)
		},
	}

	schedCmd.AddCommand(listCmd, getCmd, createCmd, updateCmd, deleteCmd)
	a.rootCmd.AddCommand(schedCmd)
}

var validFrequencies = map[string]bool{
	"never":          true,
	"daily":          true,
	"weekly":         true,
	"everyOtherWeek": true,
	"twiceAMonth":    true,
	"every4Weeks":    true,
	"monthly":        true,
	"everyOtherMonth": true,
	"every3Months":   true,
	"every4Months":   true,
	"twiceAYear":     true,
	"yearly":         true,
	"everyOtherYear": true,
}

func validateFrequency(freq string) error {
	if !validFrequencies[freq] {
		return fmt.Errorf("invalid frequency %q; valid values: never, daily, weekly, everyOtherWeek, twiceAMonth, every4Weeks, monthly, everyOtherMonth, every3Months, every4Months, twiceAYear, yearly, everyOtherYear", freq)
	}
	return nil
}
