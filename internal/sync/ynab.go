package sync

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/bwishan/ynab-cli/internal/api"
)

// SyncTransactions delta-syncs transactions from YNAB into the local store.
func SyncTransactions(ctx context.Context, client *api.Client, store *Store, planID, sinceDate, txType string) ([]byte, error) {
	lastKnowledge, err := store.GetLastKnowledge(ctx, planID)
	if err != nil {
		return nil, err
	}

	data, err := client.GetTransactions(planID, sinceDate, txType, lastKnowledge)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Data struct {
			Transactions []struct {
				ID           string `json:"id"`
				Date         string `json:"date"`
				Amount       int64  `json:"amount"`
				PayeeName    string `json:"payee_name"`
				CategoryName string `json:"category_name"`
				Memo         string `json:"memo"`
				Cleared      string `json:"cleared"`
				Approved     bool   `json:"approved"`
				AccountID    string `json:"account_id"`
				AccountName  string `json:"account_name"`
				CategoryID   string `json:"category_id"`
				PayeeID      string `json:"payee_id"`
				Deleted      bool   `json:"deleted"`
			} `json:"transactions"`
			ServerKnowledge int64 `json:"server_knowledge"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parse transactions response for sync: %w", err)
	}

	records := make([]TransactionRecord, 0, len(resp.Data.Transactions))
	for _, tx := range resp.Data.Transactions {
		raw, err := json.Marshal(tx)
		if err != nil {
			return nil, fmt.Errorf("marshal transaction %s: %w", tx.ID, err)
		}
		records = append(records, TransactionRecord{
			PlanID:       planID,
			ID:           tx.ID,
			Date:         tx.Date,
			Amount:       tx.Amount,
			PayeeName:    tx.PayeeName,
			CategoryName: tx.CategoryName,
			Memo:         tx.Memo,
			Cleared:      tx.Cleared,
			Approved:     tx.Approved,
			AccountID:    tx.AccountID,
			AccountName:  tx.AccountName,
			CategoryID:   tx.CategoryID,
			PayeeID:      tx.PayeeID,
			Deleted:      tx.Deleted,
			RawJSON:      raw,
		})
	}

	if err := store.UpsertTransactions(ctx, planID, records, resp.Data.ServerKnowledge); err != nil {
		return nil, err
	}

	return store.SearchTransactions(ctx, planID, SearchOptions{SinceDate: sinceDate, Type: txType})
}
