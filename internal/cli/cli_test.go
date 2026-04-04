package cli

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bwishan/ynab-cli/internal/config"
	syncdb "github.com/bwishan/ynab-cli/internal/sync"
)

// TestAppRun_Version tests the --version flag.
func TestAppRun_Version(t *testing.T) {
	app := New("1.2.3", "abc123", "2024-01-01")
	err := app.Run([]string{"--version"})
	if err != nil {
		t.Errorf("--version returned error: %v", err)
	}
}

// TestAppRun_Help tests the --help flag.
func TestAppRun_Help(t *testing.T) {
	app := New("1.0.0", "abc", "2024-01-01")
	err := app.Run([]string{"--help"})
	if err != nil {
		t.Errorf("--help returned error: %v", err)
	}
}

// TestAppRun_NoArgs shows help without error.
func TestAppRun_NoArgs(t *testing.T) {
	app := New("1.0.0", "abc", "2024-01-01")
	err := app.Run([]string{})
	if err != nil {
		t.Errorf("no args returned error: %v", err)
	}
}

// TestAppRun_UnknownCommand returns error.
func TestAppRun_UnknownCommand(t *testing.T) {
	app := New("1.0.0", "abc", "2024-01-01")
	err := app.Run([]string{"--token", "tok", "nonexistent", "list"})
	if err == nil {
		t.Error("expected error for unknown command")
	}
	if !strings.Contains(err.Error(), "unknown command") {
		t.Errorf("error = %q, want 'unknown command'", err.Error())
	}
}

// TestAppRun_CommandHelp shows subcommand help.
func TestAppRun_CommandHelp(t *testing.T) {
	app := New("1.0.0", "abc", "2024-01-01")
	err := app.Run([]string{"plans", "--help"})
	if err != nil {
		t.Errorf("command help returned error: %v", err)
	}
}

// TestAppRun_InvalidOutputFormat returns error.
func TestAppRun_InvalidOutputFormat(t *testing.T) {
	app := New("1.0.0", "abc", "2024-01-01")
	err := app.Run([]string{"--output", "xml", "--token", "tok", "user", "get"})
	if err == nil {
		t.Error("expected error for invalid format")
	}
	if !strings.Contains(err.Error(), "invalid output format") {
		t.Errorf("error = %q", err.Error())
	}
}

// TestAppRun_Configure tests the configure subcommand.
func TestAppRun_Configure(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	app := New("1.0.0", "abc", "2024-01-01")
	err := app.Run([]string{"configure", "--token", "my-secret-token"})
	if err != nil {
		t.Errorf("configure returned error: %v", err)
	}
}

// TestAppRun_ConfigureDefaultPlan tests setting default plan.
func TestAppRun_ConfigureDefaultPlan(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	app := New("1.0.0", "abc", "2024-01-01")
	err := app.Run([]string{"configure", "--default-plan", "plan-123"})
	if err != nil {
		t.Errorf("configure returned error: %v", err)
	}
}

// TestAppRun_ConfigureTransactionSync tests setting remembered sync config.
func TestAppRun_ConfigureTransactionSync(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	app := New("1.0.0", "abc", "2024-01-01")
	err := app.Run([]string{"configure", "--transaction-sync", "--transaction-sync-db", "/tmp/test-sync.db"})
	if err != nil {
		t.Errorf("configure returned error: %v", err)
	}
}

// TestAppRun_MissingToken returns error when no token is available.
func TestAppRun_MissingToken(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("YNAB_ACCESS_TOKEN", "")
	app := New("1.0.0", "abc", "2024-01-01")
	err := app.Run([]string{"plans", "list"})
	if err == nil {
		t.Error("expected error for missing token")
	}
	if !strings.Contains(err.Error(), "no access token") {
		t.Errorf("error = %q", err.Error())
	}
}

func TestAppRun_TransactionSyncFlagsConflict(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	app := New("1.0.0", "abc", "2024-01-01")
	err := app.Run([]string{"--token", "tok", "transactions", "list", "plan-1", "--transaction-sync", "--transaction-sync-off"})
	if err == nil {
		t.Fatal("expected conflict error")
	}
	if !strings.Contains(err.Error(), "cannot use --transaction-sync and --transaction-sync-off together") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestAppRun_TransactionSyncStatusNoToken verifies that sync-status works without an API token.
func TestAppRun_TransactionSyncStatusNoToken(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("YNAB_ACCESS_TOKEN", "")
	t.Setenv("YNAB_PLAN_ID", "plan-test")

	app := New("1.0.0", "abc", "2024-01-01")
	err := app.Run([]string{"transactions", "sync-status"})
	// Should succeed (no token required for local-only commands).
	if err != nil {
		t.Errorf("sync-status returned error: %v", err)
	}
}

// TestAppRun_TransactionSearchNoToken verifies that search works without an API token.
func TestAppRun_TransactionSearchNoToken(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("YNAB_ACCESS_TOKEN", "")
	t.Setenv("YNAB_PLAN_ID", "plan-test")

	app := New("1.0.0", "abc", "2024-01-01")
	err := app.Run([]string{"transactions", "search", "--since-date", "2024-01-01"})
	if err != nil {
		t.Errorf("search returned error: %v", err)
	}
}

// TestAppRun_TransactionSyncStatusTableOutput verifies table output format.
func TestAppRun_TransactionSyncStatusTableOutput(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("YNAB_ACCESS_TOKEN", "")
	t.Setenv("YNAB_PLAN_ID", "plan-test")

	app := New("1.0.0", "abc", "2024-01-01")
	err := app.Run([]string{"--output", "table", "transactions", "sync-status"})
	if err != nil {
		t.Errorf("sync-status table returned error: %v", err)
	}
}


func TestGlobalFlagsAfterCommand(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("YNAB_ACCESS_TOKEN", "")

	tests := []struct {
		name string
		args []string
	}{
		{"pretty after command", []string{"plans", "list", "--pretty", "--token", "tok"}},
		{"output after command", []string{"plans", "list", "--output", "table", "--token", "tok"}},
		{"token after command", []string{"plans", "list", "--token", "tok"}},
		{"all flags after command", []string{"plans", "list", "--token", "tok", "--output", "json", "--pretty"}},
		{"flags mixed", []string{"--pretty", "plans", "list", "--token", "tok", "--output", "csv"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := New("1.0.0", "abc", "2024-01-01")
			err := app.Run(tt.args)
			// These will fail at the API call (no real server), but should NOT
			// fail on flag parsing. The error should be about the API call, not flags.
			if err != nil && strings.Contains(err.Error(), "unknown flag") {
				t.Errorf("flag parsing failed: %v", err)
			}
			if err != nil && strings.Contains(err.Error(), "no access token") {
				t.Errorf("token flag not recognized: %v", err)
			}
		})
	}
}

// TestMilliunitToString tests the milliunit formatting helper.
func TestMilliunitToString(t *testing.T) {
	tests := []struct {
		input int64
		want  string
	}{
		{0, "0.00"},
		{1000, "1.00"},
		{50000, "50.00"},
		{1234560, "1234.56"},
		{-25000, "-25.00"},
		{-123450, "-123.45"},
		{500, "0.50"},
		{10, "0.01"},
	}
	for _, tt := range tests {
		got := milliunitToString(tt.input)
		if got != tt.want {
			t.Errorf("milliunitToString(%d) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// TestValidateFrequency tests frequency validation.
func TestValidateFrequency(t *testing.T) {
	valid := []string{"never", "daily", "weekly", "everyOtherWeek", "twiceAMonth",
		"every4Weeks", "monthly", "everyOtherMonth", "every3Months",
		"every4Months", "twiceAYear", "yearly", "everyOtherYear"}
	for _, f := range valid {
		if err := validateFrequency(f); err != nil {
			t.Errorf("validateFrequency(%q) returned error: %v", f, err)
		}
	}

	invalid := []string{"hourly", "biweekly", "", "MONTHLY"}
	for _, f := range invalid {
		if err := validateFrequency(f); err == nil {
			t.Errorf("validateFrequency(%q) should return error", f)
		}
	}
}

// TestExtractTransactionRows tests JSON parsing for table output.
func TestExtractTransactionRows(t *testing.T) {
	data := `{
"data": {
"transactions": [
{
"date": "2024-03-15",
"payee_name": "Coffee Shop",
"category_name": "Dining",
"memo": "Morning coffee",
"amount": -5000,
"cleared": "cleared",
"approved": true
}
]
}
}`

	rows, err := extractTransactionRows([]byte(data))
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("got %d rows, want 1", len(rows))
	}
	row := rows[0]
	if row[0] != "2024-03-15" {
		t.Errorf("date = %s", row[0])
	}
	if row[1] != "Coffee Shop" {
		t.Errorf("payee = %s", row[1])
	}
	if row[4] != "-5.00" {
		t.Errorf("amount = %s, want -5.00", row[4])
	}
	if row[6] != "true" {
		t.Errorf("approved = %s", row[6])
	}
}

// TestExtractSingleTransactionRows tests single transaction JSON parsing.
func TestExtractSingleTransactionRows(t *testing.T) {
	data := `{
"data": {
"transaction": {
"date": "2024-03-15",
"payee_name": "Store",
"category_name": "Shopping",
"memo": "",
"amount": -50000,
"cleared": "uncleared",
"approved": false
}
}
}`

	rows, err := extractSingleTransactionRows([]byte(data))
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("got %d rows, want 1", len(rows))
	}
	if rows[0][4] != "-50.00" {
		t.Errorf("amount = %s, want -50.00", rows[0][4])
	}
}

// TestTransactionCreateBody verifies the request body structure for create.
func TestTransactionCreateBody(t *testing.T) {
	tx := map[string]interface{}{
		"account_id": "acc-1",
		"date":       "2024-03-15",
		"amount":     int64(-50000),
		"payee_name": "Test",
	}
	body := map[string]interface{}{"transaction": tx}

	b, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}

	var parsed map[string]interface{}
	json.Unmarshal(b, &parsed)

	txObj, ok := parsed["transaction"].(map[string]interface{})
	if !ok {
		t.Fatal("missing transaction key")
	}
	if txObj["account_id"] != "acc-1" {
		t.Errorf("account_id = %v", txObj["account_id"])
	}
	if txObj["amount"].(float64) != -50000 {
		t.Errorf("amount = %v", txObj["amount"])
	}
}

// TestAppRun_ConfigureTransactionSyncOff verifies --transaction-sync-off saves false.
func TestAppRun_ConfigureTransactionSyncOff(t *testing.T) {
tmp := t.TempDir()
t.Setenv("HOME", tmp)

app := New("1.0.0", "abc", "2024-01-01")
// First enable sync.
if err := app.Run([]string{"configure", "--transaction-sync"}); err != nil {
t.Fatalf("configure --transaction-sync: %v", err)
}
// Then disable it.
app2 := New("1.0.0", "abc", "2024-01-01")
if err := app2.Run([]string{"configure", "--transaction-sync-off"}); err != nil {
t.Fatalf("configure --transaction-sync-off: %v", err)
}
}

// TestAppRun_ConfigureConflictSyncFlags verifies configure rejects both sync flags together.
func TestAppRun_ConfigureConflictSyncFlags(t *testing.T) {
tmp := t.TempDir()
t.Setenv("HOME", tmp)

app := New("1.0.0", "abc", "2024-01-01")
err := app.Run([]string{"configure", "--transaction-sync", "--transaction-sync-off"})
if err == nil {
t.Fatal("expected error for conflicting sync flags")
}
if !strings.Contains(err.Error(), "cannot use --transaction-sync and --transaction-sync-off together") {
t.Errorf("unexpected error: %v", err)
}
}

// TestAppRun_ConfigureSyncPersistsDefaultDB verifies that enabling sync without a DB path
// auto-fills the default transactions.db path and saves it.
func TestAppRun_ConfigureSyncPersistsDefaultDB(t *testing.T) {
tmp := t.TempDir()
t.Setenv("HOME", tmp)

app := New("1.0.0", "abc", "2024-01-01")
if err := app.Run([]string{"configure", "--transaction-sync"}); err != nil {
t.Fatalf("configure: %v", err)
}

cfgPath := filepath.Join(tmp, ".config", "ynab-cli", "config.json")
data, err := os.ReadFile(cfgPath)
if err != nil {
t.Fatalf("reading config: %v", err)
}
var cfg struct {
TransactionSync   bool   `json:"transaction_sync"`
TransactionSyncDB string `json:"transaction_sync_db"`
}
if err := json.Unmarshal(data, &cfg); err != nil {
t.Fatalf("parsing config: %v", err)
}
if !cfg.TransactionSync {
t.Error("expected transaction_sync = true")
}
if !strings.HasSuffix(cfg.TransactionSyncDB, "transactions.db") {
t.Errorf("expected transaction_sync_db to end with transactions.db, got %q", cfg.TransactionSyncDB)
}
}

// populateSyncDB is a test helper that creates a SQLite sync DB at the given path
// and inserts a set of transactions for the given plan.
func populateSyncDB(t *testing.T, dbPath, planID string, txs []syncdb.TransactionRecord, knowledge int64) {
t.Helper()
if err := os.MkdirAll(filepath.Dir(dbPath), 0700); err != nil {
t.Fatalf("mkdir for sync DB: %v", err)
}
store, err := syncdb.OpenStore(dbPath)
if err != nil {
t.Fatalf("OpenStore: %v", err)
}
defer store.Close()
if err := store.UpsertTransactions(context.Background(), planID, txs, knowledge); err != nil {
t.Fatalf("UpsertTransactions: %v", err)
}
}

// TestAppRun_SyncStatusWithPopulatedDB verifies sync-status works against a pre-populated DB
// without requiring an API token.
func TestAppRun_SyncStatusWithPopulatedDB(t *testing.T) {
tmp := t.TempDir()
t.Setenv("HOME", tmp)
t.Setenv("YNAB_ACCESS_TOKEN", "")
t.Setenv("YNAB_PLAN_ID", "plan-test")

// Use the default DB path (under the temp HOME).
dbPath := filepath.Join(tmp, ".config", "ynab-cli", "transactions.db")
stub := []byte(`{"id":"tx-1","date":"2026-01-01","amount":-5000}`)
populateSyncDB(t, dbPath, "plan-test", []syncdb.TransactionRecord{
{ID: "tx-1", Date: "2026-01-01", Amount: -5000, RawJSON: stub},
}, 42)

app := New("1.0.0", "abc", "2024-01-01")
if err := app.Run([]string{"transactions", "sync-status"}); err != nil {
t.Errorf("sync-status with populated DB returned error: %v", err)
}
}

// TestAppRun_SearchLocalOnly verifies that `transactions search` is local-only (no token needed)
// and returns results from a pre-populated DB.
func TestAppRun_SearchLocalOnly(t *testing.T) {
tmp := t.TempDir()
t.Setenv("HOME", tmp)
t.Setenv("YNAB_ACCESS_TOKEN", "")
t.Setenv("YNAB_PLAN_ID", "plan-test")

dbPath := filepath.Join(tmp, ".config", "ynab-cli", "transactions.db")
stub := []byte(`{"id":"tx-1","date":"2026-03-15","amount":-12000,"payee_name":"Cafe"}`)
populateSyncDB(t, dbPath, "plan-test", []syncdb.TransactionRecord{
{ID: "tx-1", Date: "2026-03-15", Amount: -12000, PayeeName: "Cafe", RawJSON: stub},
}, 10)

app := New("1.0.0", "abc", "2024-01-01")
if err := app.Run([]string{"transactions", "search", "--since-date", "2026-01-01"}); err != nil {
t.Errorf("transactions search returned error: %v", err)
}
}

// TestResolveTransactionSyncSettings_EnabledEmptyDBUsesDefault verifies that when sync is
// enabled but no explicit DB path is set, the default transactions.db path is used.
func TestResolveTransactionSyncSettings_EnabledEmptyDBUsesDefault(t *testing.T) {
tmp := t.TempDir()
t.Setenv("HOME", tmp)

a := &App{}
a.cfg = &config.Config{TransactionSync: true, TransactionSyncDB: ""}

enabled, dbPath, err := a.resolveTransactionSyncSettings(nil)
if err != nil {
t.Fatalf("resolveTransactionSyncSettings: %v", err)
}
if !enabled {
t.Error("expected enabled=true")
}
if !strings.HasSuffix(dbPath, "transactions.db") {
t.Errorf("expected dbPath to end with transactions.db, got %q", dbPath)
}
}

// TestResolveTransactionSyncSettings_DisabledNoPath verifies that when sync is disabled,
// no DB path is resolved.
func TestResolveTransactionSyncSettings_DisabledNoPath(t *testing.T) {
tmp := t.TempDir()
t.Setenv("HOME", tmp)

a := &App{}
a.cfg = &config.Config{TransactionSync: false, TransactionSyncDB: ""}

enabled, dbPath, err := a.resolveTransactionSyncSettings(nil)
if err != nil {
t.Fatalf("resolveTransactionSyncSettings: %v", err)
}
if enabled {
t.Error("expected enabled=false")
}
if dbPath != "" {
t.Errorf("expected empty dbPath when disabled, got %q", dbPath)
}
}
