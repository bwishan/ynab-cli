package cli

import (
	"encoding/json"
	"strings"
	"testing"
)


// TestAppRun_Version tests the --version flag.
func TestAppRun_Version(t *testing.T) {
	app := New("1.2.3", "abc123", "2024-01-01")
	// --version should not return an error
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
	// "plans" with no subcommand should show help, not error
	err := app.Run([]string{"--token", "tok", "plans"})
	if err != nil {
		t.Errorf("command help returned error: %v", err)
	}
}

// TestAppRun_CommandWithDashHelp shows subcommand help.
func TestAppRun_CommandWithDashHelp(t *testing.T) {
	app := New("1.0.0", "abc", "2024-01-01")
	err := app.Run([]string{"--token", "tok", "plans", "--help"})
	if err != nil {
		t.Errorf("command --help returned error: %v", err)
	}
}

// TestAppRun_InvalidOutputFormat returns error.
func TestAppRun_InvalidOutputFormat(t *testing.T) {
	app := New("1.0.0", "abc", "2024-01-01")
	err := app.Run([]string{"--output", "xml", "--token", "tok", "plans", "list"})
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

// TestParseFlag tests the flag parsing helper from transactions.go.
func TestParseFlag(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		idx      int
		wantKey  string
		wantVal  string
		wantCons int
		wantFlag bool
	}{
		{"positional", []string{"plan-id"}, 0, "", "", 0, false},
		{"flag_with_value", []string{"--date", "2024-01-01"}, 0, "date", "2024-01-01", 1, true},
		{"flag_with_equals", []string{"--date=2024-01-01"}, 0, "date", "2024-01-01", 0, true},
		{"boolean_flag_no_next", []string{"--approved"}, 0, "approved", "", 0, true},
		{"boolean_flag_next_is_flag", []string{"--approved", "--date"}, 0, "approved", "", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, val, consumed, isFlag := parseFlag(tt.args, tt.idx)
			if key != tt.wantKey {
				t.Errorf("key = %q, want %q", key, tt.wantKey)
			}
			if val != tt.wantVal {
				t.Errorf("val = %q, want %q", val, tt.wantVal)
			}
			if consumed != tt.wantCons {
				t.Errorf("consumed = %d, want %d", consumed, tt.wantCons)
			}
			if isFlag != tt.wantFlag {
				t.Errorf("isFlag = %v, want %v", isFlag, tt.wantFlag)
			}
		})
	}
}

// TestParseTransactionFilterArgs tests the transaction filter argument parser.
func TestParseTransactionFilterArgs(t *testing.T) {
	args := []string{"plan-1", "--since-date", "2024-01-01", "--type", "uncategorized", "--last-knowledge", "500"}
	positional, sinceDate, txType, lastKnowledge, err := parseTransactionFilterArgs(args)
	if err != nil {
		t.Fatal(err)
	}
	if len(positional) != 1 || positional[0] != "plan-1" {
		t.Errorf("positional = %v, want [plan-1]", positional)
	}
	if sinceDate != "2024-01-01" {
		t.Errorf("sinceDate = %s", sinceDate)
	}
	if txType != "uncategorized" {
		t.Errorf("txType = %s", txType)
	}
	if lastKnowledge != 500 {
		t.Errorf("lastKnowledge = %d", lastKnowledge)
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

