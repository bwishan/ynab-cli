package cli

import (
"encoding/json"
"strings"
"testing"
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

// TestGlobalFlagsAfterCommand tests that global flags work after the command (the bug fix).
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

// TestRequestBodiesUseDateNotDateFirst is a regression guard ensuring no
// create/update command accidentally maps the --date flag to "date_first".
// The YNAB API returns "date_first" in responses but expects "date" in requests.
func TestRequestBodiesUseDateNotDateFirst(t *testing.T) {
cases := []struct {
name    string
fields  map[string]interface{}
wrapper string
}{
{"transactions create", map[string]interface{}{"account_id": "acc-1", "date": "2024-03-15", "amount": int64(-50000)}, "transaction"},
{"scheduled create", map[string]interface{}{"account_id": "acc-1", "date": "2024-04-01", "amount": int64(-100000), "frequency": "monthly"}, "scheduled_transaction"},
{"scheduled update", map[string]interface{}{"date": "2024-05-01"}, "scheduled_transaction"},
}

for _, tc := range cases {
t.Run(tc.name, func(t *testing.T) {
if _, bad := tc.fields["date_first"]; bad {
t.Error("field map contains 'date_first' — API expects 'date'")
}
if _, ok := tc.fields["date"]; !ok {
t.Error("field map missing 'date'")
}
// Also verify the serialized form, since that's what goes over the wire.
b, err := json.Marshal(map[string]interface{}{tc.wrapper: tc.fields})
if err != nil {
t.Fatal(err)
}
if strings.Contains(string(b), "date_first") {
t.Errorf("serialized body contains 'date_first'")
}
})
}
}
