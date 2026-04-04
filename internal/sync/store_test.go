package sync

import (
	"context"
	"database/sql"
	"encoding/json"
	"path/filepath"
	"testing"
)

func TestStoreRoundTrip(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "transactions.db")
	store, err := OpenStore(path)
	if err != nil {
		t.Fatalf("OpenStore: %v", err)
	}
	defer store.Close()

	raw1 := []byte(`{"id":"tx-1","date":"2026-04-01","amount":-12345,"payee_name":"Coffee","category_name":"Dining","memo":"latte","cleared":"cleared","approved":true,"account_id":"acc-1","account_name":"Checking","category_id":"cat-1","payee_id":"pay-1","deleted":false}`)
	raw2 := []byte(`{"id":"tx-2","date":"2026-04-02","amount":-5000,"payee_name":"Store","category_name":"","memo":"snacks","cleared":"uncleared","approved":false,"account_id":"acc-1","account_name":"Checking","category_id":"","payee_id":"pay-2","deleted":false}`)

	err = store.UpsertTransactions(ctx, "plan-1", []TransactionRecord{
		{ID: "tx-1", Date: "2026-04-01", Amount: -12345, PayeeName: "Coffee", CategoryName: "Dining", Memo: "latte", Cleared: "cleared", Approved: true, AccountID: "acc-1", AccountName: "Checking", CategoryID: "cat-1", PayeeID: "pay-1", RawJSON: raw1},
		{ID: "tx-2", Date: "2026-04-02", Amount: -5000, PayeeName: "Store", Memo: "snacks", Cleared: "uncleared", Approved: false, AccountID: "acc-1", AccountName: "Checking", CategoryID: "", PayeeID: "pay-2", RawJSON: raw2},
	}, 42)
	if err != nil {
		t.Fatalf("UpsertTransactions: %v", err)
	}

	knowledge, err := store.GetLastKnowledge(ctx, "plan-1")
	if err != nil {
		t.Fatalf("GetLastKnowledge: %v", err)
	}
	if knowledge != 42 {
		t.Fatalf("knowledge = %d, want 42", knowledge)
	}

	data, err := store.SearchTransactions(ctx, "plan-1", SearchOptions{Type: "uncategorized"})
	if err != nil {
		t.Fatalf("SearchTransactions: %v", err)
	}
	var parsed struct {
		Data struct {
			Transactions    []map[string]interface{} `json:"transactions"`
			ServerKnowledge int64                    `json:"server_knowledge"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Unmarshal search response: %v", err)
	}
	if len(parsed.Data.Transactions) != 1 {
		t.Fatalf("len(transactions) = %d, want 1", len(parsed.Data.Transactions))
	}
	if parsed.Data.Transactions[0]["id"] != "tx-2" {
		t.Fatalf("got transaction id %v, want tx-2", parsed.Data.Transactions[0]["id"])
	}
	if parsed.Data.ServerKnowledge != 42 {
		t.Fatalf("server_knowledge = %d, want 42", parsed.Data.ServerKnowledge)
	}

	single, err := store.GetTransaction(ctx, "plan-1", "tx-1")
	if err != nil {
		t.Fatalf("GetTransaction: %v", err)
	}
	var singleParsed struct {
		Data struct {
			Transaction map[string]interface{} `json:"transaction"`
		} `json:"data"`
	}
	if err := json.Unmarshal(single, &singleParsed); err != nil {
		t.Fatalf("Unmarshal single response: %v", err)
	}
	if singleParsed.Data.Transaction["id"] != "tx-1" {
		t.Fatalf("single id = %v, want tx-1", singleParsed.Data.Transaction["id"])
	}
}

func TestStoreGetTransactionNotFound(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "transactions.db")
	store, err := OpenStore(path)
	if err != nil {
		t.Fatalf("OpenStore: %v", err)
	}
	defer store.Close()

	_, err = store.GetTransaction(ctx, "plan-1", "missing")
	if err != sql.ErrNoRows {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestStoreStatusCountsDeletedAndLastKnowledge(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "transactions.db")
	store, err := OpenStore(path)
	if err != nil {
		t.Fatalf("OpenStore: %v", err)
	}
	defer store.Close()

	rawLive := []byte(`{"id":"tx-live","date":"2026-04-01","amount":-1000,"payee_name":"Cafe","category_name":"Food","memo":"","cleared":"cleared","approved":true,"account_id":"acc-1","account_name":"Checking","category_id":"cat-1","payee_id":"pay-1","deleted":false}`)
	rawDeleted := []byte(`{"id":"tx-deleted","date":"2026-04-02","amount":-2000,"payee_name":"Shop","category_name":"Stuff","memo":"","cleared":"cleared","approved":true,"account_id":"acc-1","account_name":"Checking","category_id":"cat-2","payee_id":"pay-2","deleted":true}`)

	err = store.UpsertTransactions(ctx, "plan-1", []TransactionRecord{
		{ID: "tx-live", Date: "2026-04-01", Amount: -1000, PayeeName: "Cafe", CategoryName: "Food", Cleared: "cleared", Approved: true, AccountID: "acc-1", AccountName: "Checking", CategoryID: "cat-1", PayeeID: "pay-1", RawJSON: rawLive},
		{ID: "tx-deleted", Date: "2026-04-02", Amount: -2000, PayeeName: "Shop", CategoryName: "Stuff", Cleared: "cleared", Approved: true, AccountID: "acc-1", AccountName: "Checking", CategoryID: "cat-2", PayeeID: "pay-2", Deleted: true, RawJSON: rawDeleted},
	}, 99)
	if err != nil {
		t.Fatalf("UpsertTransactions: %v", err)
	}

	status, err := store.Status(ctx, "plan-1", path)
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if status.LastKnowledge != 99 {
		t.Fatalf("LastKnowledge = %d, want 99", status.LastKnowledge)
	}
	if status.TransactionCount != 1 {
		t.Fatalf("TransactionCount = %d, want 1", status.TransactionCount)
	}
	if status.DeletedCount != 1 {
		t.Fatalf("DeletedCount = %d, want 1", status.DeletedCount)
	}
	if status.DBPath != path {
		t.Fatalf("DBPath = %q, want %q", status.DBPath, path)
	}
	if status.LastSyncAt.IsZero() {
		t.Fatal("expected LastSyncAt to be set")
	}
}

func TestStoreSearchTransactions_FiltersByMemoAndDate(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "transactions.db")
	store, err := OpenStore(path)
	if err != nil {
		t.Fatalf("OpenStore: %v", err)
	}
	defer store.Close()

	raw1 := []byte(`{"id":"tx-1","date":"2026-04-01","amount":-1000,"payee_name":"Cafe","category_name":"Food","memo":"coffee beans","cleared":"cleared","approved":true,"account_id":"acc-1","account_name":"Checking","category_id":"cat-1","payee_id":"pay-1","deleted":false}`)
	raw2 := []byte(`{"id":"tx-2","date":"2026-04-03","amount":-2000,"payee_name":"Market","category_name":"Groceries","memo":"vegetables","cleared":"cleared","approved":true,"account_id":"acc-1","account_name":"Checking","category_id":"cat-2","payee_id":"pay-2","deleted":false}`)

	err = store.UpsertTransactions(ctx, "plan-1", []TransactionRecord{
		{ID: "tx-1", Date: "2026-04-01", Amount: -1000, PayeeName: "Cafe", CategoryName: "Food", Memo: "coffee beans", Cleared: "cleared", Approved: true, AccountID: "acc-1", AccountName: "Checking", CategoryID: "cat-1", PayeeID: "pay-1", RawJSON: raw1},
		{ID: "tx-2", Date: "2026-04-03", Amount: -2000, PayeeName: "Market", CategoryName: "Groceries", Memo: "vegetables", Cleared: "cleared", Approved: true, AccountID: "acc-1", AccountName: "Checking", CategoryID: "cat-2", PayeeID: "pay-2", RawJSON: raw2},
	}, 7)
	if err != nil {
		t.Fatalf("UpsertTransactions: %v", err)
	}

	data, err := store.SearchTransactions(ctx, "plan-1", SearchOptions{SinceDate: "2026-04-02", Memo: "veg"})
	if err != nil {
		t.Fatalf("SearchTransactions: %v", err)
	}
	var parsed struct {
		Data struct {
			Transactions []map[string]interface{} `json:"transactions"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Unmarshal search response: %v", err)
	}
	if len(parsed.Data.Transactions) != 1 {
		t.Fatalf("len(transactions) = %d, want 1", len(parsed.Data.Transactions))
	}
	if parsed.Data.Transactions[0]["id"] != "tx-2" {
		t.Fatalf("id = %v, want tx-2", parsed.Data.Transactions[0]["id"])
	}
}
