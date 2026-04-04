package cli

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/bwishan/ynab-cli/internal/config"
	syncdb "github.com/bwishan/ynab-cli/internal/sync"
)

func (a *App) registerTransactionCommands() {
	txCmd := &cobra.Command{
		Use:   "transactions",
		Short: "Manage transactions",
	}

	// ── list ────────────────────────────────────────────────────────────
	listCmd := &cobra.Command{
		Use:   "list [plan-id]",
		Short: "List all transactions for a plan",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.validateTransactionSyncFlags(cmd); err != nil {
				return err
			}
			var planIDArg string
			if len(args) > 0 {
				planIDArg = args[0]
			}
			planID, err := config.ResolvePlan(planIDArg)
			if err != nil {
				return err
			}

			sinceDate, _ := cmd.Flags().GetString("since-date")
			txType, _ := cmd.Flags().GetString("type")
			lastKnowledge, _ := cmd.Flags().GetInt64("last-knowledge")

			useSync, dbPath, err := a.resolveTransactionSyncSettings(cmd)
			if err != nil {
				return err
			}
			if useSync {
				store, err := syncdb.OpenStore(dbPath)
				if err != nil {
					return err
				}
				defer store.Close()

				data, err := syncdb.SyncTransactions(context.Background(), a.client, store, planID, sinceDate, txType)
				if err != nil {
					return err
				}
				return a.printer.PrintResult(data, transactionHeaders, extractTransactionRows)
			}

			data, err := a.client.GetTransactions(planID, sinceDate, txType, lastKnowledge)
			if err != nil {
				return err
			}
			return a.printer.PrintResult(data, transactionHeaders, extractTransactionRows)
		},
	}
	listCmd.Flags().String("since-date", "", "Only return transactions on or after this date (ISO 8601)")
	listCmd.Flags().String("type", "", "Transaction type filter")
	listCmd.Flags().Int64("last-knowledge", 0, "Server knowledge for delta sync")
	a.addTransactionSyncFlags(listCmd)

	// ── get ─────────────────────────────────────────────────────────────
	getCmd := &cobra.Command{
		Use:   "get <plan-id> <transaction-id>",
		Short: "Get a specific transaction",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.validateTransactionSyncFlags(cmd); err != nil {
				return err
			}
			planID, err := config.ResolvePlan(args[0])
			if err != nil {
				return err
			}
			useSync, dbPath, err := a.resolveTransactionSyncSettings(cmd)
			if err != nil {
				return err
			}
			if useSync {
				store, err := syncdb.OpenStore(dbPath)
				if err != nil {
					return err
				}
				defer store.Close()

				data, err := store.GetTransaction(context.Background(), planID, args[1])
				if err == nil {
					return a.printer.PrintResult(data, transactionHeaders, extractSingleTransactionRows)
				}
				if err != sql.ErrNoRows {
					return err
				}

				if _, err := syncdb.SyncTransactions(context.Background(), a.client, store, planID, "", ""); err != nil {
					return err
				}
				data, err = store.GetTransaction(context.Background(), planID, args[1])
				if err != nil {
					if err == sql.ErrNoRows {
						return fmt.Errorf("transaction %s not found in sync cache after refresh", args[1])
					}
					return err
				}
				return a.printer.PrintResult(data, transactionHeaders, extractSingleTransactionRows)
			}
			data, err := a.client.GetTransactionByID(planID, args[1])
			if err != nil {
				return err
			}
			return a.printer.PrintResult(data, transactionHeaders, extractSingleTransactionRows)
		},
	}

	// ── create ──────────────────────────────────────────────────────────
	createCmd := &cobra.Command{
		Use:   "create [plan-id]",
		Short: "Create a new transaction",
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

			accountID, _ := cmd.Flags().GetString("account-id")
			date, _ := cmd.Flags().GetString("date")
			amount, _ := cmd.Flags().GetInt64("amount")
			payeeID, _ := cmd.Flags().GetString("payee-id")
			payeeName, _ := cmd.Flags().GetString("payee-name")
			categoryID, _ := cmd.Flags().GetString("category-id")
			memo, _ := cmd.Flags().GetString("memo")
			cleared, _ := cmd.Flags().GetString("cleared")
			approved, _ := cmd.Flags().GetBool("approved")
			flagColor, _ := cmd.Flags().GetString("flag-color")
			importID, _ := cmd.Flags().GetString("import-id")

			tx := map[string]interface{}{
				"account_id": accountID,
				"date":       date,
				"amount":     amount,
			}
			if payeeID != "" {
				tx["payee_id"] = payeeID
			}
			if payeeName != "" {
				tx["payee_name"] = payeeName
			}
			if categoryID != "" {
				tx["category_id"] = categoryID
			}
			if memo != "" {
				tx["memo"] = memo
			}
			if cleared != "" {
				tx["cleared"] = cleared
			}
			if approved {
				tx["approved"] = true
			}
			if flagColor != "" {
				tx["flag_color"] = flagColor
			}
			if importID != "" {
				tx["import_id"] = importID
			}

			body := map[string]interface{}{
				"transaction": tx,
			}

			data, err := a.client.CreateTransaction(planID, body)
			if err != nil {
				return err
			}
			return a.printer.PrintResult(data, nil, nil)
		},
	}
	createCmd.Flags().String("account-id", "", "Account ID (required)")
	createCmd.Flags().String("date", "", "Transaction date (required)")
	createCmd.Flags().Int64("amount", 0, "Amount in milliunits (required)")
	createCmd.Flags().String("payee-id", "", "Payee ID")
	createCmd.Flags().String("payee-name", "", "Payee name")
	createCmd.Flags().String("category-id", "", "Category ID")
	createCmd.Flags().String("memo", "", "Memo")
	createCmd.Flags().String("cleared", "", "Cleared status: cleared, uncleared, reconciled")
	createCmd.Flags().Bool("approved", false, "Whether the transaction is approved")
	createCmd.Flags().String("flag-color", "", "Flag color")
	createCmd.Flags().String("import-id", "", "Import ID")
	_ = createCmd.MarkFlagRequired("account-id")
	_ = createCmd.MarkFlagRequired("date")
	_ = createCmd.MarkFlagRequired("amount")

	// ── update ──────────────────────────────────────────────────────────
	updateCmd := &cobra.Command{
		Use:   "update <plan-id> <transaction-id>",
		Short: "Update a transaction",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			planID, err := config.ResolvePlan(args[0])
			if err != nil {
				return err
			}
			transactionID := args[1]

			accountID, _ := cmd.Flags().GetString("account-id")
			date, _ := cmd.Flags().GetString("date")
			amount, _ := cmd.Flags().GetInt64("amount")
			payeeID, _ := cmd.Flags().GetString("payee-id")
			payeeName, _ := cmd.Flags().GetString("payee-name")
			categoryID, _ := cmd.Flags().GetString("category-id")
			memo, _ := cmd.Flags().GetString("memo")
			cleared, _ := cmd.Flags().GetString("cleared")
			approved, _ := cmd.Flags().GetBool("approved")
			flagColor, _ := cmd.Flags().GetString("flag-color")
			importID, _ := cmd.Flags().GetString("import-id")

			tx := map[string]interface{}{}
			if cmd.Flags().Changed("account-id") {
				tx["account_id"] = accountID
			}
			if cmd.Flags().Changed("date") {
				tx["date"] = date
			}
			if cmd.Flags().Changed("amount") {
				tx["amount"] = amount
			}
			if cmd.Flags().Changed("payee-id") {
				tx["payee_id"] = payeeID
			}
			if cmd.Flags().Changed("payee-name") {
				tx["payee_name"] = payeeName
			}
			if cmd.Flags().Changed("category-id") {
				tx["category_id"] = categoryID
			}
			if cmd.Flags().Changed("memo") {
				tx["memo"] = memo
			}
			if cmd.Flags().Changed("cleared") {
				tx["cleared"] = cleared
			}
			if approved {
				tx["approved"] = true
			}
			if cmd.Flags().Changed("flag-color") {
				tx["flag_color"] = flagColor
			}
			if cmd.Flags().Changed("import-id") {
				tx["import_id"] = importID
			}

			if len(tx) == 0 {
				return fmt.Errorf("at least one field flag is required")
			}

			body := map[string]interface{}{
				"transaction": tx,
			}

			data, err := a.client.UpdateTransaction(planID, transactionID, body)
			if err != nil {
				return err
			}
			return a.printer.PrintResult(data, nil, nil)
		},
	}
	updateCmd.Flags().String("account-id", "", "Account ID")
	updateCmd.Flags().String("date", "", "Transaction date")
	updateCmd.Flags().Int64("amount", 0, "Amount in milliunits")
	updateCmd.Flags().String("payee-id", "", "Payee ID")
	updateCmd.Flags().String("payee-name", "", "Payee name")
	updateCmd.Flags().String("category-id", "", "Category ID")
	updateCmd.Flags().String("memo", "", "Memo")
	updateCmd.Flags().String("cleared", "", "Cleared status: cleared, uncleared, reconciled")
	updateCmd.Flags().Bool("approved", false, "Whether the transaction is approved")
	updateCmd.Flags().String("flag-color", "", "Flag color")
	updateCmd.Flags().String("import-id", "", "Import ID")

	// ── delete ──────────────────────────────────────────────────────────
	deleteCmd := &cobra.Command{
		Use:   "delete <plan-id> <transaction-id>",
		Short: "Delete a transaction",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			planID, err := config.ResolvePlan(args[0])
			if err != nil {
				return err
			}
			data, err := a.client.DeleteTransaction(planID, args[1])
			if err != nil {
				return err
			}
			return a.printer.PrintResult(data, nil, nil)
		},
	}

	// ── import ──────────────────────────────────────────────────────────
	importCmd := &cobra.Command{
		Use:   "import [plan-id]",
		Short: "Import transactions for linked accounts",
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
			data, err := a.client.ImportTransactions(planID)
			if err != nil {
				return err
			}
			return a.printer.PrintResult(data, nil, nil)
		},
	}

	// ── list-by-account ─────────────────────────────────────────────────
	listByAccountCmd := &cobra.Command{
		Use:   "list-by-account <plan-id> <account-id>",
		Short: "List transactions for a specific account",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			planID, err := config.ResolvePlan(args[0])
			if err != nil {
				return err
			}
			sinceDate, _ := cmd.Flags().GetString("since-date")
			txType, _ := cmd.Flags().GetString("type")
			lastKnowledge, _ := cmd.Flags().GetInt64("last-knowledge")

			data, err := a.client.GetTransactionsByAccount(planID, args[1], sinceDate, txType, lastKnowledge)
			if err != nil {
				return err
			}
			return a.printer.PrintResult(data, transactionHeaders, extractTransactionRows)
		},
	}
	listByAccountCmd.Flags().String("since-date", "", "Only return transactions on or after this date (ISO 8601)")
	listByAccountCmd.Flags().String("type", "", "Transaction type filter")
	listByAccountCmd.Flags().Int64("last-knowledge", 0, "Server knowledge for delta sync")

	// ── list-by-category ────────────────────────────────────────────────
	listByCategoryCmd := &cobra.Command{
		Use:   "list-by-category <plan-id> <category-id>",
		Short: "List transactions for a specific category",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			planID, err := config.ResolvePlan(args[0])
			if err != nil {
				return err
			}
			sinceDate, _ := cmd.Flags().GetString("since-date")
			txType, _ := cmd.Flags().GetString("type")
			lastKnowledge, _ := cmd.Flags().GetInt64("last-knowledge")

			data, err := a.client.GetTransactionsByCategory(planID, args[1], sinceDate, txType, lastKnowledge)
			if err != nil {
				return err
			}
			return a.printer.PrintResult(data, transactionHeaders, extractTransactionRows)
		},
	}
	listByCategoryCmd.Flags().String("since-date", "", "Only return transactions on or after this date (ISO 8601)")
	listByCategoryCmd.Flags().String("type", "", "Transaction type filter")
	listByCategoryCmd.Flags().Int64("last-knowledge", 0, "Server knowledge for delta sync")

	// ── list-by-payee ───────────────────────────────────────────────────
	listByPayeeCmd := &cobra.Command{
		Use:   "list-by-payee <plan-id> <payee-id>",
		Short: "List transactions for a specific payee",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			planID, err := config.ResolvePlan(args[0])
			if err != nil {
				return err
			}
			sinceDate, _ := cmd.Flags().GetString("since-date")
			txType, _ := cmd.Flags().GetString("type")
			lastKnowledge, _ := cmd.Flags().GetInt64("last-knowledge")

			data, err := a.client.GetTransactionsByPayee(planID, args[1], sinceDate, txType, lastKnowledge)
			if err != nil {
				return err
			}
			return a.printer.PrintResult(data, transactionHeaders, extractTransactionRows)
		},
	}
	listByPayeeCmd.Flags().String("since-date", "", "Only return transactions on or after this date (ISO 8601)")
	listByPayeeCmd.Flags().String("type", "", "Transaction type filter")
	listByPayeeCmd.Flags().Int64("last-knowledge", 0, "Server knowledge for delta sync")

	// ── list-by-month ───────────────────────────────────────────────────
	listByMonthCmd := &cobra.Command{
		Use:   "list-by-month <plan-id> <month>",
		Short: "List transactions for a specific month",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			planID, err := config.ResolvePlan(args[0])
			if err != nil {
				return err
			}
			sinceDate, _ := cmd.Flags().GetString("since-date")
			txType, _ := cmd.Flags().GetString("type")
			lastKnowledge, _ := cmd.Flags().GetInt64("last-knowledge")

			data, err := a.client.GetTransactionsByMonth(planID, args[1], sinceDate, txType, lastKnowledge)
			if err != nil {
				return err
			}
			return a.printer.PrintResult(data, transactionHeaders, extractTransactionRows)
		},
	}
	listByMonthCmd.Flags().String("since-date", "", "Only return transactions on or after this date (ISO 8601)")
	listByMonthCmd.Flags().String("type", "", "Transaction type filter")
	listByMonthCmd.Flags().Int64("last-knowledge", 0, "Server knowledge for delta sync")

	// ── sync ────────────────────────────────────────────────────────────
	syncCmd := &cobra.Command{
		Use:   "sync [plan-id]",
		Short: "Delta-sync transactions into the local SQLite cache",
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
			dbPath, err := a.requireTransactionSyncDB(cmd)
			if err != nil {
				return err
			}
			store, err := syncdb.OpenStore(dbPath)
			if err != nil {
				return err
			}
			defer store.Close()

			sinceDate, _ := cmd.Flags().GetString("since-date")
			txType, _ := cmd.Flags().GetString("type")
			data, err := syncdb.SyncTransactions(context.Background(), a.client, store, planID, sinceDate, txType)
			if err != nil {
				return err
			}
			return a.printer.PrintResult(data, transactionHeaders, extractTransactionRows)
		},
	}
	syncCmd.Flags().String("since-date", "", "Only sync transactions on or after this date (ISO 8601)")
	syncCmd.Flags().String("type", "", "Transaction type filter")
	a.addTransactionSyncFlags(syncCmd)

	// ── sync-status ─────────────────────────────────────────────────────
	syncStatusCmd := &cobra.Command{
		Use:         "sync-status [plan-id]",
		Short:       "Show transaction sync/cache status for a plan",
		Args:        cobra.MaximumNArgs(1),
		Annotations: map[string]string{"localOnly": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			var planIDArg string
			if len(args) > 0 {
				planIDArg = args[0]
			}
			planID, err := config.ResolvePlan(planIDArg)
			if err != nil {
				return err
			}
			dbPath, err := a.requireTransactionSyncDB(cmd)
			if err != nil {
				return err
			}
			store, err := syncdb.OpenStore(dbPath)
			if err != nil {
				return err
			}
			defer store.Close()

			status, err := store.Status(context.Background(), planID, dbPath)
			if err != nil {
				return err
			}
			payload, err := json.Marshal(struct {
				Data *syncdb.SyncStatus `json:"data"`
			}{Data: status})
			if err != nil {
				return err
			}
			headers := []string{"PLAN_ID", "DB_PATH", "LAST_KNOWLEDGE", "LAST_SYNC_AT", "TRANSACTIONS", "DELETED"}
			extractRows := func(data []byte) ([][]string, error) {
				var resp struct {
					Data syncdb.SyncStatus `json:"data"`
				}
				if err := json.Unmarshal(data, &resp); err != nil {
					return nil, err
				}
				return [][]string{{
					resp.Data.PlanID,
					resp.Data.DBPath,
					strconv.FormatInt(resp.Data.LastKnowledge, 10),
					resp.Data.LastSyncAt.Format("2006-01-02T15:04:05Z07:00"),
					strconv.Itoa(resp.Data.TransactionCount),
					strconv.Itoa(resp.Data.DeletedCount),
				}}, nil
			}
			return a.printer.PrintResult(payload, headers, extractRows)
		},
	}

	// ── search ──────────────────────────────────────────────────────────
	searchCmd := &cobra.Command{
		Use:         "search [plan-id]",
		Short:       "Search cached transactions in the local SQLite sync DB",
		Args:        cobra.MaximumNArgs(1),
		Annotations: map[string]string{"localOnly": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			var planIDArg string
			if len(args) > 0 {
				planIDArg = args[0]
			}
			planID, err := config.ResolvePlan(planIDArg)
			if err != nil {
				return err
			}
			dbPath, err := a.requireTransactionSyncDB(cmd)
			if err != nil {
				return err
			}
			store, err := syncdb.OpenStore(dbPath)
			if err != nil {
				return err
			}
			defer store.Close()

			sinceDate, _ := cmd.Flags().GetString("since-date")
			beforeDate, _ := cmd.Flags().GetString("before-date")
			txType, _ := cmd.Flags().GetString("type")
			accountID, _ := cmd.Flags().GetString("account-id")
			categoryID, _ := cmd.Flags().GetString("category-id")
			payeeID, _ := cmd.Flags().GetString("payee-id")
			memo, _ := cmd.Flags().GetString("memo")
			limit, _ := cmd.Flags().GetInt("limit")

			data, err := store.SearchTransactions(context.Background(), planID, syncdb.SearchOptions{
				SinceDate:  sinceDate,
				BeforeDate: beforeDate,
				Type:       txType,
				AccountID:  accountID,
				CategoryID: categoryID,
				PayeeID:    payeeID,
				Memo:       memo,
				Limit:      limit,
			})
			if err != nil {
				return err
			}
			return a.printer.PrintResult(data, transactionHeaders, extractTransactionRows)
		},
	}
	searchCmd.Flags().String("since-date", "", "Only return transactions on or after this date (ISO 8601)")
	searchCmd.Flags().String("before-date", "", "Only return transactions on or before this date (ISO 8601)")
	searchCmd.Flags().String("type", "", "Transaction type filter")
	searchCmd.Flags().String("account-id", "", "Filter by account ID")
	searchCmd.Flags().String("category-id", "", "Filter by category ID")
	searchCmd.Flags().String("payee-id", "", "Filter by payee ID")
	searchCmd.Flags().String("memo", "", "Filter memo substring")
	searchCmd.Flags().Int("limit", 100, "Maximum number of cached transactions to return")
	a.addTransactionSyncFlags(searchCmd)

	// ── bulk-update ─────────────────────────────────────────────────────
	bulkUpdateCmd := &cobra.Command{
		Use:   "bulk-update [plan-id]",
		Short: "Update multiple transactions at once",
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

			dataStr, _ := cmd.Flags().GetString("data")

			var txArray json.RawMessage
			if err := json.Unmarshal([]byte(dataStr), &txArray); err != nil {
				return fmt.Errorf("invalid --data JSON: %w", err)
			}

			body := map[string]json.RawMessage{
				"transactions": txArray,
			}

			data, err := a.client.UpdateTransactions(planID, body)
			if err != nil {
				return err
			}
			return a.printer.PrintResult(data, nil, nil)
		},
	}
	bulkUpdateCmd.Flags().String("data", "", "JSON string of transactions to update (required)")
	_ = bulkUpdateCmd.MarkFlagRequired("data")

	txCmd.AddCommand(listCmd, getCmd, createCmd, updateCmd, deleteCmd, importCmd,
		syncCmd, syncStatusCmd, searchCmd,
		listByAccountCmd, listByCategoryCmd, listByPayeeCmd, listByMonthCmd, bulkUpdateCmd)
	a.rootCmd.AddCommand(txCmd)
}

// transactionHeaders defines the table columns for transaction list output.
var transactionHeaders = []string{"DATE", "PAYEE", "CATEGORY", "MEMO", "AMOUNT", "CLEARED", "APPROVED"}

// extractTransactionRows parses a multi-transaction response for table output.
func extractTransactionRows(data []byte) ([][]string, error) {
	var resp struct {
		Data struct {
			Transactions []struct {
				Date         string `json:"date"`
				PayeeName    string `json:"payee_name"`
				CategoryName string `json:"category_name"`
				Memo         string `json:"memo"`
				Amount       int64  `json:"amount"`
				Cleared      string `json:"cleared"`
				Approved     bool   `json:"approved"`
			} `json:"transactions"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing transactions response: %w", err)
	}

	var rows [][]string
	for _, t := range resp.Data.Transactions {
		rows = append(rows, []string{
			t.Date,
			t.PayeeName,
			t.CategoryName,
			t.Memo,
			milliunitToString(t.Amount),
			t.Cleared,
			strconv.FormatBool(t.Approved),
		})
	}
	return rows, nil
}

// extractSingleTransactionRows parses a single-transaction response for table output.
func extractSingleTransactionRows(data []byte) ([][]string, error) {
	var resp struct {
		Data struct {
			Transaction struct {
				Date         string `json:"date"`
				PayeeName    string `json:"payee_name"`
				CategoryName string `json:"category_name"`
				Memo         string `json:"memo"`
				Amount       int64  `json:"amount"`
				Cleared      string `json:"cleared"`
				Approved     bool   `json:"approved"`
			} `json:"transaction"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing transaction response: %w", err)
	}

	t := resp.Data.Transaction
	return [][]string{{
		t.Date,
		t.PayeeName,
		t.CategoryName,
		t.Memo,
		milliunitToString(t.Amount),
		t.Cleared,
		strconv.FormatBool(t.Approved),
	}}, nil
}
