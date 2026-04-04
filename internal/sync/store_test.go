package sync

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
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

func TestOpenStore_EmptyPath(t *testing.T) {
	_, err := OpenStore("")
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestOpenStore_CreatesDirectory(t *testing.T) {
	// OpenStore should create the parent directory if it doesn't exist.
	path := filepath.Join(t.TempDir(), "nested", "dir", "transactions.db")
	store, err := OpenStore(path)
	if err != nil {
		t.Fatalf("OpenStore: %v", err)
	}
	store.Close()
}

func TestStoreStatus_Empty(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "transactions.db")
	store, err := OpenStore(path)
	if err != nil {
		t.Fatalf("OpenStore: %v", err)
	}
	defer store.Close()

	status, err := store.Status(ctx, "plan-1", path)
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if status.PlanID != "plan-1" {
		t.Errorf("PlanID = %q, want plan-1", status.PlanID)
	}
	if status.DBPath != path {
		t.Errorf("DBPath = %q, want %q", status.DBPath, path)
	}
	if status.LastKnowledge != 0 {
		t.Errorf("LastKnowledge = %d, want 0", status.LastKnowledge)
	}
	if status.TransactionCount != 0 {
		t.Errorf("TransactionCount = %d, want 0", status.TransactionCount)
	}
	if status.DeletedCount != 0 {
		t.Errorf("DeletedCount = %d, want 0", status.DeletedCount)
	}
}

func TestStoreStatus_WithData(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "transactions.db")
	store, err := OpenStore(path)
	if err != nil {
		t.Fatalf("OpenStore: %v", err)
	}
	defer store.Close()

	stub := []byte(`{"id":"x","date":"2026-01-01","amount":-1000}`)
	err = store.UpsertTransactions(ctx, "plan-1", []TransactionRecord{
		{ID: "tx-a", Date: "2026-01-01", Amount: -1000, RawJSON: stub},
		{ID: "tx-b", Date: "2026-01-02", Amount: -2000, RawJSON: stub},
		{ID: "tx-del", Date: "2026-01-03", Amount: -3000, Deleted: true, RawJSON: stub},
	}, 99)
	if err != nil {
		t.Fatalf("UpsertTransactions: %v", err)
	}

	status, err := store.Status(ctx, "plan-1", path)
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if status.LastKnowledge != 99 {
		t.Errorf("LastKnowledge = %d, want 99", status.LastKnowledge)
	}
	if status.TransactionCount != 2 {
		t.Errorf("TransactionCount = %d, want 2", status.TransactionCount)
	}
	if status.DeletedCount != 1 {
		t.Errorf("DeletedCount = %d, want 1", status.DeletedCount)
	}
	if status.LastSyncAt.IsZero() {
		t.Error("expected LastSyncAt to be non-zero after sync")
	}
}

func TestGetLastKnowledge_UnseenPlan(t *testing.T) {
	ctx := context.Background()
	store, err := OpenStore(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatalf("OpenStore: %v", err)
	}
	defer store.Close()

	k, err := store.GetLastKnowledge(ctx, "never-synced")
	if err != nil {
		t.Fatalf("GetLastKnowledge: %v", err)
	}
	if k != 0 {
		t.Errorf("expected 0 for unseen plan, got %d", k)
	}
}

func TestUpsertTransactions_UpdatesExisting(t *testing.T) {
	ctx := context.Background()
	store, err := OpenStore(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatalf("OpenStore: %v", err)
	}
	defer store.Close()

	rawV1 := []byte(`{"id":"tx-1","payee_name":"OldPayee"}`)
	rawV2 := []byte(`{"id":"tx-1","payee_name":"NewPayee"}`)

	if err := store.UpsertTransactions(ctx, "plan-1", []TransactionRecord{
		{ID: "tx-1", Date: "2026-01-01", PayeeName: "OldPayee", RawJSON: rawV1},
	}, 1); err != nil {
		t.Fatalf("first upsert: %v", err)
	}
	if err := store.UpsertTransactions(ctx, "plan-1", []TransactionRecord{
		{ID: "tx-1", Date: "2026-01-01", PayeeName: "NewPayee", RawJSON: rawV2},
	}, 2); err != nil {
		t.Fatalf("second upsert: %v", err)
	}

	data, err := store.GetTransaction(ctx, "plan-1", "tx-1")
	if err != nil {
		t.Fatalf("GetTransaction: %v", err)
	}
	var resp struct {
		Data struct {
			Transaction map[string]interface{} `json:"transaction"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if resp.Data.Transaction["payee_name"] != "NewPayee" {
		t.Errorf("payee_name = %v, want NewPayee", resp.Data.Transaction["payee_name"])
	}

	k, _ := store.GetLastKnowledge(ctx, "plan-1")
	if k != 2 {
		t.Errorf("last_knowledge = %d, want 2 after second upsert", k)
	}
}

func TestGetTransaction_DeletedReturnsNotFound(t *testing.T) {
	ctx := context.Background()
	store, err := OpenStore(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatalf("OpenStore: %v", err)
	}
	defer store.Close()

	raw := []byte(`{"id":"tx-del","date":"2026-01-01","amount":-1000,"deleted":true}`)
	if err := store.UpsertTransactions(ctx, "plan-1", []TransactionRecord{
		{ID: "tx-del", Date: "2026-01-01", Deleted: true, RawJSON: raw},
	}, 1); err != nil {
		t.Fatalf("UpsertTransactions: %v", err)
	}

	_, err = store.GetTransaction(ctx, "plan-1", "tx-del")
	if err != sql.ErrNoRows {
		t.Fatalf("expected sql.ErrNoRows for deleted transaction, got %v", err)
	}
}

// newStoreWithTxs is a test helper that opens a store and upserts a set of transactions.
func newStoreWithTxs(t *testing.T, planID string, records []TransactionRecord, knowledge int64) *Store {
	t.Helper()
	store, err := OpenStore(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatalf("OpenStore: %v", err)
	}
	t.Cleanup(func() { store.Close() })
	if err := store.UpsertTransactions(context.Background(), planID, records, knowledge); err != nil {
		t.Fatalf("UpsertTransactions: %v", err)
	}
	return store
}

func txIDs(t *testing.T, data []byte) []string {
	t.Helper()
	var resp struct {
		Data struct {
			Transactions []map[string]interface{} `json:"transactions"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		t.Fatalf("txIDs unmarshal: %v", err)
	}
	ids := make([]string, len(resp.Data.Transactions))
	for i, tx := range resp.Data.Transactions {
		ids[i], _ = tx["id"].(string)
	}
	return ids
}

func TestSearchTransactions_SinceDateFilter(t *testing.T) {
	ctx := context.Background()
	raw := func(id, date string) []byte {
		return []byte(`{"id":"` + id + `","date":"` + date + `","amount":-1000}`)
	}
	store := newStoreWithTxs(t, "p", []TransactionRecord{
		{ID: "old", Date: "2026-01-15", RawJSON: raw("old", "2026-01-15")},
		{ID: "new", Date: "2026-04-01", RawJSON: raw("new", "2026-04-01")},
	}, 1)

	data, err := store.SearchTransactions(ctx, "p", SearchOptions{SinceDate: "2026-03-01"})
	if err != nil {
		t.Fatalf("SearchTransactions: %v", err)
	}
	ids := txIDs(t, data)
	if len(ids) != 1 || ids[0] != "new" {
		t.Errorf("expected [new], got %v", ids)
	}
}

func TestSearchTransactions_BeforeDateFilter(t *testing.T) {
	ctx := context.Background()
	raw := func(id, date string) []byte {
		return []byte(`{"id":"` + id + `","date":"` + date + `","amount":-1000}`)
	}
	store := newStoreWithTxs(t, "p", []TransactionRecord{
		{ID: "old", Date: "2026-01-15", RawJSON: raw("old", "2026-01-15")},
		{ID: "new", Date: "2026-04-01", RawJSON: raw("new", "2026-04-01")},
	}, 1)

	data, err := store.SearchTransactions(ctx, "p", SearchOptions{BeforeDate: "2026-02-01"})
	if err != nil {
		t.Fatalf("SearchTransactions: %v", err)
	}
	ids := txIDs(t, data)
	if len(ids) != 1 || ids[0] != "old" {
		t.Errorf("expected [old], got %v", ids)
	}
}

func TestSearchTransactions_AccountIDFilter(t *testing.T) {
	ctx := context.Background()
	raw := func(id string) []byte {
		return []byte(`{"id":"` + id + `","date":"2026-01-01","amount":-1000}`)
	}
	store := newStoreWithTxs(t, "p", []TransactionRecord{
		{ID: "acc-a-tx", Date: "2026-01-01", AccountID: "acc-a", RawJSON: raw("acc-a-tx")},
		{ID: "acc-b-tx", Date: "2026-01-01", AccountID: "acc-b", RawJSON: raw("acc-b-tx")},
	}, 1)

	data, err := store.SearchTransactions(ctx, "p", SearchOptions{AccountID: "acc-b"})
	if err != nil {
		t.Fatalf("SearchTransactions: %v", err)
	}
	ids := txIDs(t, data)
	if len(ids) != 1 || ids[0] != "acc-b-tx" {
		t.Errorf("expected [acc-b-tx], got %v", ids)
	}
}

func TestSearchTransactions_MemoSubstring(t *testing.T) {
	ctx := context.Background()
	raw := func(id, memo string) []byte {
		return []byte(`{"id":"` + id + `","date":"2026-01-01","amount":-1000,"memo":"` + memo + `"}`)
	}
	store := newStoreWithTxs(t, "p", []TransactionRecord{
		{ID: "t1", Date: "2026-01-01", Memo: "coffee latte", RawJSON: raw("t1", "coffee latte")},
		{ID: "t2", Date: "2026-01-01", Memo: "groceries", RawJSON: raw("t2", "groceries")},
	}, 1)

	data, err := store.SearchTransactions(ctx, "p", SearchOptions{Memo: "latte"})
	if err != nil {
		t.Fatalf("SearchTransactions: %v", err)
	}
	ids := txIDs(t, data)
	if len(ids) != 1 || ids[0] != "t1" {
		t.Errorf("expected [t1], got %v", ids)
	}
}

func TestSearchTransactions_UnapprovedType(t *testing.T) {
	ctx := context.Background()
	raw := func(id string, approved bool) []byte {
		a := "false"
		if approved {
			a = "true"
		}
		return []byte(`{"id":"` + id + `","date":"2026-01-01","amount":-1000,"approved":` + a + `}`)
	}
	store := newStoreWithTxs(t, "p", []TransactionRecord{
		{ID: "approved-tx", Date: "2026-01-01", Approved: true, RawJSON: raw("approved-tx", true)},
		{ID: "pending-tx", Date: "2026-01-02", Approved: false, RawJSON: raw("pending-tx", false)},
	}, 1)

	data, err := store.SearchTransactions(ctx, "p", SearchOptions{Type: "unapproved"})
	if err != nil {
		t.Fatalf("SearchTransactions: %v", err)
	}
	ids := txIDs(t, data)
	if len(ids) != 1 || ids[0] != "pending-tx" {
		t.Errorf("expected [pending-tx], got %v", ids)
	}
}

func TestSearchTransactions_Limit(t *testing.T) {
	ctx := context.Background()
	records := make([]TransactionRecord, 5)
	for i := range records {
		id := fmt.Sprintf("tx-%d", i)
		records[i] = TransactionRecord{
			ID:      id,
			Date:    "2026-01-01",
			RawJSON: []byte(`{"id":"` + id + `","date":"2026-01-01","amount":-1000}`),
		}
	}
	store := newStoreWithTxs(t, "p", records, 1)

	data, err := store.SearchTransactions(ctx, "p", SearchOptions{Limit: 3})
	if err != nil {
		t.Fatalf("SearchTransactions: %v", err)
	}
	ids := txIDs(t, data)
	if len(ids) != 3 {
		t.Errorf("expected 3 results with Limit=3, got %d", len(ids))
	}
}

func TestSearchTransactions_PayeeIDFilter(t *testing.T) {
	ctx := context.Background()
	raw := func(id string) []byte {
		return []byte(`{"id":"` + id + `","date":"2026-01-01","amount":-1000}`)
	}
	store := newStoreWithTxs(t, "p", []TransactionRecord{
		{ID: "t1", Date: "2026-01-01", PayeeID: "pay-x", RawJSON: raw("t1")},
		{ID: "t2", Date: "2026-01-01", PayeeID: "pay-y", RawJSON: raw("t2")},
	}, 1)

	data, err := store.SearchTransactions(ctx, "p", SearchOptions{PayeeID: "pay-x"})
	if err != nil {
		t.Fatalf("SearchTransactions: %v", err)
	}
	ids := txIDs(t, data)
	if len(ids) != 1 || ids[0] != "t1" {
		t.Errorf("expected [t1], got %v", ids)
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
