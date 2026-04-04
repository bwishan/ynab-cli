package sync

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// Store provides a SQLite-backed local cache for transactions.
type Store struct {
	db *sql.DB
}

// TransactionRecord stores the subset of transaction fields needed for querying/output,
// plus the raw JSON returned by YNAB.
type TransactionRecord struct {
	PlanID       string
	ID           string
	Date         string
	Amount       int64
	PayeeName    string
	CategoryName string
	Memo         string
	Cleared      string
	Approved     bool
	AccountID    string
	AccountName  string
	CategoryID   string
	PayeeID      string
	Deleted      bool
	RawJSON      []byte
	UpdatedAt    time.Time
}

// SearchOptions controls cached transaction querying.
type SearchOptions struct {
	SinceDate  string
	BeforeDate string
	Type       string
	AccountID  string
	CategoryID string
	PayeeID    string
	Memo       string
	Limit      int
}

// SyncStatus describes the current sync state for a plan.
type SyncStatus struct {
	PlanID           string    `json:"plan_id"`
	DBPath           string    `json:"db_path"`
	LastKnowledge    int64     `json:"last_knowledge"`
	LastSyncAt       time.Time `json:"last_sync_at,omitempty"`
	TransactionCount int       `json:"transaction_count"`
	DeletedCount     int       `json:"deleted_count"`
}

// OpenStore opens or creates the SQLite cache database.
func OpenStore(path string) (*Store, error) {
	if strings.TrimSpace(path) == "" {
		return nil, fmt.Errorf("sync db path is required")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, fmt.Errorf("creating sync db directory: %w", err)
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("opening sqlite db: %w", err)
	}

	store := &Store{db: db}
	if err := store.init(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) init() error {
	stmts := []string{
		`PRAGMA journal_mode=WAL;`,
		`CREATE TABLE IF NOT EXISTS transactions (
			plan_id TEXT NOT NULL,
			id TEXT NOT NULL,
			date TEXT NOT NULL,
			amount INTEGER NOT NULL,
			payee_name TEXT,
			category_name TEXT,
			memo TEXT,
			cleared TEXT,
			approved INTEGER NOT NULL,
			account_id TEXT,
			account_name TEXT,
			category_id TEXT,
			payee_id TEXT,
			deleted INTEGER NOT NULL DEFAULT 0,
			raw_json TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			PRIMARY KEY (plan_id, id)
		);`,
		`CREATE INDEX IF NOT EXISTS idx_transactions_plan_date ON transactions(plan_id, date DESC);`,
		`CREATE INDEX IF NOT EXISTS idx_transactions_plan_account ON transactions(plan_id, account_id);`,
		`CREATE INDEX IF NOT EXISTS idx_transactions_plan_category ON transactions(plan_id, category_id);`,
		`CREATE INDEX IF NOT EXISTS idx_transactions_plan_payee ON transactions(plan_id, payee_id);`,
		`CREATE TABLE IF NOT EXISTS sync_state (
			plan_id TEXT PRIMARY KEY,
			last_knowledge INTEGER NOT NULL DEFAULT 0,
			last_sync_at TEXT
		);`,
	}

	for _, stmt := range stmts {
		if _, err := s.db.Exec(stmt); err != nil {
			return fmt.Errorf("initializing sync db: %w", err)
		}
	}
	return nil
}

// Close closes the underlying DB handle.
func (s *Store) Close() error {
	return s.db.Close()
}

// UpsertTransactions upserts a batch of transactions and updates sync metadata.
func (s *Store) UpsertTransactions(ctx context.Context, planID string, txs []TransactionRecord, lastKnowledge int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO transactions (
			plan_id, id, date, amount, payee_name, category_name, memo, cleared,
			approved, account_id, account_name, category_id, payee_id, deleted, raw_json, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(plan_id, id) DO UPDATE SET
			date=excluded.date,
			amount=excluded.amount,
			payee_name=excluded.payee_name,
			category_name=excluded.category_name,
			memo=excluded.memo,
			cleared=excluded.cleared,
			approved=excluded.approved,
			account_id=excluded.account_id,
			account_name=excluded.account_name,
			category_id=excluded.category_id,
			payee_id=excluded.payee_id,
			deleted=excluded.deleted,
			raw_json=excluded.raw_json,
			updated_at=excluded.updated_at;
	`)
	if err != nil {
		return fmt.Errorf("prepare upsert: %w", err)
	}
	defer stmt.Close()

	now := time.Now().UTC().Format(time.RFC3339Nano)
	for _, rec := range txs {
		if _, err = stmt.ExecContext(ctx,
			planID,
			rec.ID,
			rec.Date,
			rec.Amount,
			rec.PayeeName,
			rec.CategoryName,
			rec.Memo,
			rec.Cleared,
			boolToInt(rec.Approved),
			rec.AccountID,
			rec.AccountName,
			rec.CategoryID,
			rec.PayeeID,
			boolToInt(rec.Deleted),
			string(rec.RawJSON),
			now,
		); err != nil {
			return fmt.Errorf("upsert transaction %s: %w", rec.ID, err)
		}
	}

	if _, err = tx.ExecContext(ctx, `
		INSERT INTO sync_state(plan_id, last_knowledge, last_sync_at)
		VALUES (?, ?, ?)
		ON CONFLICT(plan_id) DO UPDATE SET
			last_knowledge=excluded.last_knowledge,
			last_sync_at=excluded.last_sync_at;
	`, planID, lastKnowledge, now); err != nil {
		return fmt.Errorf("update sync state: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

// GetLastKnowledge returns the last server knowledge seen for a plan.
func (s *Store) GetLastKnowledge(ctx context.Context, planID string) (int64, error) {
	var knowledge int64
	err := s.db.QueryRowContext(ctx, `SELECT last_knowledge FROM sync_state WHERE plan_id = ?`, planID).Scan(&knowledge)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("query sync state: %w", err)
	}
	return knowledge, nil
}

// GetTransaction returns a single cached transaction.
func (s *Store) GetTransaction(ctx context.Context, planID, transactionID string) ([]byte, error) {
	var raw string
	err := s.db.QueryRowContext(ctx, `
		SELECT raw_json FROM transactions
		WHERE plan_id = ? AND id = ? AND deleted = 0
	`, planID, transactionID).Scan(&raw)
	if err == sql.ErrNoRows {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, fmt.Errorf("query transaction: %w", err)
	}

	return wrapSingleTransactionJSON([]byte(raw))
}

// SearchTransactions queries cached transactions and returns a YNAB-like JSON response.
func (s *Store) SearchTransactions(ctx context.Context, planID string, opts SearchOptions) ([]byte, error) {
	query := `
		SELECT raw_json FROM transactions
		WHERE plan_id = ? AND deleted = 0
	`
	args := []interface{}{planID}

	if opts.SinceDate != "" {
		query += ` AND date >= ?`
		args = append(args, opts.SinceDate)
	}
	if opts.BeforeDate != "" {
		query += ` AND date <= ?`
		args = append(args, opts.BeforeDate)
	}
	if opts.AccountID != "" {
		query += ` AND account_id = ?`
		args = append(args, opts.AccountID)
	}
	if opts.CategoryID != "" {
		query += ` AND category_id = ?`
		args = append(args, opts.CategoryID)
	}
	if opts.PayeeID != "" {
		query += ` AND payee_id = ?`
		args = append(args, opts.PayeeID)
	}
	if opts.Memo != "" {
		query += ` AND memo LIKE ?`
		args = append(args, "%"+opts.Memo+"%")
	}
	if opts.Type == "uncategorized" {
		query += ` AND (category_id = '' OR category_id IS NULL)`
	}
	if opts.Type == "unapproved" {
		query += ` AND approved = 0`
	}

	query += ` ORDER BY date DESC, id DESC`
	if opts.Limit > 0 {
		query += ` LIMIT ?`
		args = append(args, opts.Limit)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("search transactions: %w", err)
	}
	defer rows.Close()

	var transactions []json.RawMessage
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			return nil, fmt.Errorf("scan transaction row: %w", err)
		}
		transactions = append(transactions, json.RawMessage(raw))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate transaction rows: %w", err)
	}

	knowledge, err := s.GetLastKnowledge(ctx, planID)
	if err != nil {
		return nil, err
	}
	return wrapTransactionsJSON(transactions, knowledge)
}

// Status returns sync metadata and row counts.
func (s *Store) Status(ctx context.Context, planID, dbPath string) (*SyncStatus, error) {
	status := &SyncStatus{PlanID: planID, DBPath: dbPath}
	row := s.db.QueryRowContext(ctx, `SELECT last_knowledge, COALESCE(last_sync_at, '') FROM sync_state WHERE plan_id = ?`, planID)
	var lastSync string
	if err := row.Scan(&status.LastKnowledge, &lastSync); err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("query sync status: %w", err)
	}
	if lastSync != "" {
		if parsed, err := time.Parse(time.RFC3339Nano, lastSync); err == nil {
			status.LastSyncAt = parsed
		}
	}

	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM transactions WHERE plan_id = ? AND deleted = 0`, planID).Scan(&status.TransactionCount); err != nil {
		return nil, fmt.Errorf("count transactions: %w", err)
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM transactions WHERE plan_id = ? AND deleted = 1`, planID).Scan(&status.DeletedCount); err != nil {
		return nil, fmt.Errorf("count deleted transactions: %w", err)
	}

	return status, nil
}

func wrapTransactionsJSON(items []json.RawMessage, serverKnowledge int64) ([]byte, error) {
	payload := struct {
		Data struct {
			Transactions    []json.RawMessage `json:"transactions"`
			ServerKnowledge int64             `json:"server_knowledge"`
		} `json:"data"`
	}{}
	payload.Data.Transactions = items
	payload.Data.ServerKnowledge = serverKnowledge
	return json.Marshal(payload)
}

func wrapSingleTransactionJSON(item []byte) ([]byte, error) {
	payload := struct {
		Data struct {
			Transaction json.RawMessage `json:"transaction"`
		} `json:"data"`
	}{}
	payload.Data.Transaction = item
	return json.Marshal(payload)
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
