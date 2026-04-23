package cli

import (
	"strings"
	"testing"
)

func TestParseSplitFlags_JSONArray(t *testing.T) {
	splits, total, err := parseSplitFlags([]string{
		`[{"amount":-129123,"category_id":"abc"},{"amount":-24809,"category_id":"def","memo":"Dining"}]`,
	})
	if err != nil {
		t.Fatalf("parseSplitFlags returned error: %v", err)
	}
	if len(splits) != 2 {
		t.Fatalf("expected 2 splits, got %d", len(splits))
	}
	if total != -153932 {
		t.Fatalf("total = %d, want -153932", total)
	}
	if splits[0]["category_id"] != "abc" {
		t.Fatalf("first split category_id = %v, want abc", splits[0]["category_id"])
	}
}

func TestParseSplitFlags_RepeatablePairs(t *testing.T) {
	splits, total, err := parseSplitFlags([]string{
		"amount=-129123,category_id=abc,memo=Club dues",
		"amount=-24809,category_id=def,memo=Dining",
	})
	if err != nil {
		t.Fatalf("parseSplitFlags returned error: %v", err)
	}
	if len(splits) != 2 {
		t.Fatalf("expected 2 splits, got %d", len(splits))
	}
	if total != -153932 {
		t.Fatalf("total = %d, want -153932", total)
	}
	if splits[1]["memo"] != "Dining" {
		t.Fatalf("second split memo = %v, want Dining", splits[1]["memo"])
	}
}

func TestParseSplitFlags_MissingAmount(t *testing.T) {
	_, _, err := parseSplitFlags([]string{"category_id=abc,memo=No amount"})
	if err == nil {
		t.Fatal("expected error for missing amount")
	}
	if !strings.Contains(err.Error(), "missing required amount") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParseSplitFlags_UnknownKey(t *testing.T) {
	_, _, err := parseSplitFlags([]string{"amount=-100,category=abc"})
	if err == nil {
		t.Fatal("expected error for unknown key")
	}
	if !strings.Contains(err.Error(), "unknown key") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTransactionsCreateSplitCategoryValidation(t *testing.T) {
	app := New("1.0.0", "abc", "2024-01-01")
	err := app.Run([]string{
		"--token", "tok",
		"transactions", "create", "plan-1",
		"--account-id", "acc-1",
		"--date", "2024-03-15",
		"--amount", "-1000",
		"--category-id", "cat-1",
		"--split", "amount=-1000,category_id=cat-a",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "--category-id cannot be used when --split is set") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTransactionsCreateSplitAmountValidation(t *testing.T) {
	app := New("1.0.0", "abc", "2024-01-01")
	err := app.Run([]string{
		"--token", "tok",
		"transactions", "create", "plan-1",
		"--account-id", "acc-1",
		"--date", "2024-03-15",
		"--amount", "-2000",
		"--split", "amount=-1000,category_id=cat-a",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "split amounts sum to") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTransactionsUpdateSplitRequiresAmount(t *testing.T) {
	app := New("1.0.0", "abc", "2024-01-01")
	err := app.Run([]string{
		"--token", "tok",
		"transactions", "update", "plan-1", "tx-1",
		"--split", "amount=-1000,category_id=cat-a",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "--amount is required when using --split on update") {
		t.Fatalf("unexpected error: %v", err)
	}
}
