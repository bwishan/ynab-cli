package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// requestLog records details of an incoming request.
type requestLog struct {
	Method string
	Path   string
	Query  string
	Body   string
}

func newRecordingServer(t *testing.T) (*httptest.Server, *requestLog) {
	t.Helper()
	log := &requestLog{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Method = r.Method
		log.Path = r.URL.Path
		log.Query = r.URL.RawQuery
		if r.Body != nil {
			b, _ := io.ReadAll(r.Body)
			log.Body = string(b)
		}
		w.WriteHeader(200)
		w.Write([]byte(`{"data":{}}`))
	}))
	return ts, log
}

func clientFor(ts *httptest.Server) *Client {
	c := NewClient("test-token")
	c.baseURL = ts.URL
	return c
}

// ---------------------------------------------------------------------------
// User
// ---------------------------------------------------------------------------

func TestGetUser(t *testing.T) {
	ts, log := newRecordingServer(t)
	defer ts.Close()
	c := clientFor(ts)

	_, err := c.GetUser()
	if err != nil {
		t.Fatal(err)
	}
	if log.Method != "GET" {
		t.Errorf("method = %s, want GET", log.Method)
	}
	if log.Path != "/user" {
		t.Errorf("path = %s, want /user", log.Path)
	}
	if log.Body != "" {
		t.Errorf("body = %q, want empty", log.Body)
	}
}

// ---------------------------------------------------------------------------
// Plans (Budgets)
// ---------------------------------------------------------------------------

func TestGetPlans(t *testing.T) {
	t.Run("without_accounts", func(t *testing.T) {
		ts, log := newRecordingServer(t)
		defer ts.Close()
		c := clientFor(ts)

		_, err := c.GetPlans(false)
		if err != nil {
			t.Fatal(err)
		}
		if log.Path != "/budgets" {
			t.Errorf("path = %s, want /budgets", log.Path)
		}
		if log.Query != "" {
			t.Errorf("query = %s, want empty", log.Query)
		}
	})

	t.Run("with_accounts", func(t *testing.T) {
		ts, log := newRecordingServer(t)
		defer ts.Close()
		c := clientFor(ts)

		_, err := c.GetPlans(true)
		if err != nil {
			t.Fatal(err)
		}
		if log.Query != "include_accounts=true" {
			t.Errorf("query = %s, want include_accounts=true", log.Query)
		}
	})
}

func TestGetPlanByID(t *testing.T) {
	t.Run("without_knowledge", func(t *testing.T) {
		ts, log := newRecordingServer(t)
		defer ts.Close()
		c := clientFor(ts)

		_, err := c.GetPlanByID("plan-123", 0)
		if err != nil {
			t.Fatal(err)
		}
		if log.Path != "/budgets/plan-123" {
			t.Errorf("path = %s, want /budgets/plan-123", log.Path)
		}
		if log.Query != "" {
			t.Errorf("query = %s, want empty", log.Query)
		}
	})

	t.Run("with_knowledge", func(t *testing.T) {
		ts, log := newRecordingServer(t)
		defer ts.Close()
		c := clientFor(ts)

		_, err := c.GetPlanByID("plan-123", 500)
		if err != nil {
			t.Fatal(err)
		}
		if log.Query != "last_knowledge_of_server=500" {
			t.Errorf("query = %s, want last_knowledge_of_server=500", log.Query)
		}
	})
}

func TestGetPlanSettings(t *testing.T) {
	ts, log := newRecordingServer(t)
	defer ts.Close()
	c := clientFor(ts)

	_, err := c.GetPlanSettings("plan-123")
	if err != nil {
		t.Fatal(err)
	}
	if log.Path != "/budgets/plan-123/settings" {
		t.Errorf("path = %s, want /budgets/plan-123/settings", log.Path)
	}
}

// ---------------------------------------------------------------------------
// Accounts
// ---------------------------------------------------------------------------

func TestGetAccounts(t *testing.T) {
	t.Run("no_knowledge", func(t *testing.T) {
		ts, log := newRecordingServer(t)
		defer ts.Close()
		c := clientFor(ts)

		_, err := c.GetAccounts("plan-1", 0)
		if err != nil {
			t.Fatal(err)
		}
		if log.Path != "/budgets/plan-1/accounts" {
			t.Errorf("path = %s", log.Path)
		}
		if log.Query != "" {
			t.Errorf("query = %s, want empty", log.Query)
		}
	})

	t.Run("with_knowledge", func(t *testing.T) {
		ts, log := newRecordingServer(t)
		defer ts.Close()
		c := clientFor(ts)

		_, err := c.GetAccounts("plan-1", 100)
		if err != nil {
			t.Fatal(err)
		}
		if log.Path != "/budgets/plan-1/accounts" {
			t.Errorf("path = %s", log.Path)
		}
		if log.Query != "last_knowledge_of_server=100" {
			t.Errorf("query = %s", log.Query)
		}
	})
}

func TestGetAccountByID(t *testing.T) {
	ts, log := newRecordingServer(t)
	defer ts.Close()
	c := clientFor(ts)

	_, err := c.GetAccountByID("plan-1", "acc-1")
	if err != nil {
		t.Fatal(err)
	}
	if log.Path != "/budgets/plan-1/accounts/acc-1" {
		t.Errorf("path = %s", log.Path)
	}
}

func TestCreateAccount(t *testing.T) {
	ts, log := newRecordingServer(t)
	defer ts.Close()
	c := clientFor(ts)

	body := map[string]interface{}{"account": map[string]interface{}{"name": "Checking"}}
	_, err := c.CreateAccount("plan-1", body)
	if err != nil {
		t.Fatal(err)
	}
	if log.Method != "POST" {
		t.Errorf("method = %s, want POST", log.Method)
	}
	if log.Path != "/budgets/plan-1/accounts" {
		t.Errorf("path = %s", log.Path)
	}
	var reqBody map[string]interface{}
	json.Unmarshal([]byte(log.Body), &reqBody)
	if reqBody["account"] == nil {
		t.Error("body missing 'account' key")
	}
}

// ---------------------------------------------------------------------------
// Categories
// ---------------------------------------------------------------------------

func TestGetCategories(t *testing.T) {
	t.Run("no_knowledge", func(t *testing.T) {
		ts, log := newRecordingServer(t)
		defer ts.Close()
		c := clientFor(ts)

		_, err := c.GetCategories("plan-1", 0)
		if err != nil {
			t.Fatal(err)
		}
		if log.Path != "/budgets/plan-1/categories" {
			t.Errorf("path = %s, want /budgets/plan-1/categories", log.Path)
		}
		if log.Query != "" {
			t.Errorf("query = %s, want empty", log.Query)
		}
	})

	t.Run("with_knowledge", func(t *testing.T) {
		ts, log := newRecordingServer(t)
		defer ts.Close()
		c := clientFor(ts)

		_, err := c.GetCategories("plan-1", 88)
		if err != nil {
			t.Fatal(err)
		}
		if log.Path != "/budgets/plan-1/categories" {
			t.Errorf("path = %s, want /budgets/plan-1/categories", log.Path)
		}
		if log.Query != "last_knowledge_of_server=88" {
			t.Errorf("query = %s, want last_knowledge_of_server=88", log.Query)
		}
	})
}

func TestGetCategoryByID(t *testing.T) {
	ts, log := newRecordingServer(t)
	defer ts.Close()
	c := clientFor(ts)

	_, err := c.GetCategoryByID("plan-1", "cat-1")
	if err != nil {
		t.Fatal(err)
	}
	if log.Method != "GET" {
		t.Errorf("method = %s, want GET", log.Method)
	}
	if log.Path != "/budgets/plan-1/categories/cat-1" {
		t.Errorf("path = %s, want /budgets/plan-1/categories/cat-1", log.Path)
	}
}

func TestCreateCategory(t *testing.T) {
	ts, log := newRecordingServer(t)
	defer ts.Close()
	c := clientFor(ts)

	body := map[string]interface{}{"category": map[string]interface{}{"name": "Groceries"}}
	_, err := c.CreateCategory("plan-1", body)
	if err != nil {
		t.Fatal(err)
	}
	if log.Method != "POST" {
		t.Errorf("method = %s, want POST", log.Method)
	}
	if log.Path != "/budgets/plan-1/categories" {
		t.Errorf("path = %s, want /budgets/plan-1/categories", log.Path)
	}
	if log.Body == "" {
		t.Error("expected non-empty body")
	}
}

func TestUpdateCategory(t *testing.T) {
	ts, log := newRecordingServer(t)
	defer ts.Close()
	c := clientFor(ts)

	body := map[string]interface{}{"category": map[string]interface{}{"name": "Rent"}}
	_, err := c.UpdateCategory("plan-1", "cat-1", body)
	if err != nil {
		t.Fatal(err)
	}
	if log.Method != "PATCH" {
		t.Errorf("method = %s, want PATCH", log.Method)
	}
	if log.Path != "/budgets/plan-1/categories/cat-1" {
		t.Errorf("path = %s, want /budgets/plan-1/categories/cat-1", log.Path)
	}
}

func TestGetMonthCategory(t *testing.T) {
	ts, log := newRecordingServer(t)
	defer ts.Close()
	c := clientFor(ts)

	_, err := c.GetMonthCategory("plan-1", "2024-03-01", "cat-1")
	if err != nil {
		t.Fatal(err)
	}
	if log.Method != "GET" {
		t.Errorf("method = %s, want GET", log.Method)
	}
	if log.Path != "/budgets/plan-1/months/2024-03-01/categories/cat-1" {
		t.Errorf("path = %s, want /budgets/plan-1/months/2024-03-01/categories/cat-1", log.Path)
	}
}

func TestUpdateMonthCategory(t *testing.T) {
	ts, log := newRecordingServer(t)
	defer ts.Close()
	c := clientFor(ts)

	body := map[string]interface{}{"category": map[string]interface{}{"budgeted": 500000}}
	_, err := c.UpdateMonthCategory("plan-1", "2024-03-01", "cat-1", body)
	if err != nil {
		t.Fatal(err)
	}
	if log.Method != "PATCH" {
		t.Errorf("method = %s, want PATCH", log.Method)
	}
	if log.Path != "/budgets/plan-1/months/2024-03-01/categories/cat-1" {
		t.Errorf("path = %s, want /budgets/plan-1/months/2024-03-01/categories/cat-1", log.Path)
	}
	if log.Body == "" {
		t.Error("expected non-empty body")
	}
}

func TestCreateCategoryGroup(t *testing.T) {
	ts, log := newRecordingServer(t)
	defer ts.Close()
	c := clientFor(ts)

	body := map[string]interface{}{"category_group": map[string]interface{}{"name": "Bills"}}
	_, err := c.CreateCategoryGroup("plan-1", body)
	if err != nil {
		t.Fatal(err)
	}
	if log.Method != "POST" {
		t.Errorf("method = %s, want POST", log.Method)
	}
	if log.Path != "/budgets/plan-1/category_groups" {
		t.Errorf("path = %s, want /budgets/plan-1/category_groups", log.Path)
	}
}

func TestUpdateCategoryGroup(t *testing.T) {
	ts, log := newRecordingServer(t)
	defer ts.Close()
	c := clientFor(ts)

	body := map[string]interface{}{"category_group": map[string]interface{}{"name": "Utilities"}}
	_, err := c.UpdateCategoryGroup("plan-1", "grp-1", body)
	if err != nil {
		t.Fatal(err)
	}
	if log.Method != "PATCH" {
		t.Errorf("method = %s, want PATCH", log.Method)
	}
	if log.Path != "/budgets/plan-1/category_groups/grp-1" {
		t.Errorf("path = %s, want /budgets/plan-1/category_groups/grp-1", log.Path)
	}
	if log.Body == "" {
		t.Error("expected non-empty body")
	}
}

// ---------------------------------------------------------------------------
// Transactions
// ---------------------------------------------------------------------------

func TestGetTransactions_Filters(t *testing.T) {
	tests := []struct {
		name      string
		sinceDate string
		txType    string
		lk        int64
		wantQuery string
	}{
		{"no_filters", "", "", 0, ""},
		{"since_date_only", "2024-01-01", "", 0, "since_date=2024-01-01"},
		{"type_only", "", "uncategorized", 0, "type=uncategorized"},
		{"knowledge_only", "", "", 99, "last_knowledge_of_server=99"},
		{"all_filters", "2024-01-01", "unapproved", 100, "last_knowledge_of_server=100&since_date=2024-01-01&type=unapproved"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, log := newRecordingServer(t)
			defer ts.Close()
			c := clientFor(ts)

			_, err := c.GetTransactions("plan-1", tt.sinceDate, tt.txType, tt.lk)
			if err != nil {
				t.Fatal(err)
			}
			if log.Method != "GET" {
				t.Errorf("method = %s, want GET", log.Method)
			}
			if log.Path != "/budgets/plan-1/transactions" {
				t.Errorf("path = %s, want /budgets/plan-1/transactions", log.Path)
			}
			if log.Query != tt.wantQuery {
				t.Errorf("query = %s, want %s", log.Query, tt.wantQuery)
			}
		})
	}
}

func TestGetTransactionByID(t *testing.T) {
	ts, log := newRecordingServer(t)
	defer ts.Close()
	c := clientFor(ts)

	_, err := c.GetTransactionByID("plan-1", "tx-42")
	if err != nil {
		t.Fatal(err)
	}
	if log.Method != "GET" {
		t.Errorf("method = %s, want GET", log.Method)
	}
	if log.Path != "/budgets/plan-1/transactions/tx-42" {
		t.Errorf("path = %s, want /budgets/plan-1/transactions/tx-42", log.Path)
	}
}

func TestCreateTransaction(t *testing.T) {
	ts, log := newRecordingServer(t)
	defer ts.Close()
	c := clientFor(ts)

	body := map[string]interface{}{"transaction": map[string]interface{}{"amount": -50000}}
	_, err := c.CreateTransaction("plan-1", body)
	if err != nil {
		t.Fatal(err)
	}
	if log.Method != "POST" {
		t.Errorf("method = %s, want POST", log.Method)
	}
	if log.Path != "/budgets/plan-1/transactions" {
		t.Errorf("path = %s", log.Path)
	}
	if log.Body == "" {
		t.Error("expected non-empty body")
	}
}

func TestUpdateTransactions(t *testing.T) {
	ts, log := newRecordingServer(t)
	defer ts.Close()
	c := clientFor(ts)

	body := map[string]interface{}{
		"transactions": []map[string]interface{}{
			{"id": "tx-1", "memo": "updated"},
			{"id": "tx-2", "memo": "also updated"},
		},
	}
	_, err := c.UpdateTransactions("plan-1", body)
	if err != nil {
		t.Fatal(err)
	}
	if log.Method != "PATCH" {
		t.Errorf("method = %s, want PATCH", log.Method)
	}
	if log.Path != "/budgets/plan-1/transactions" {
		t.Errorf("path = %s, want /budgets/plan-1/transactions", log.Path)
	}
	if log.Body == "" {
		t.Error("expected non-empty body")
	}
}

func TestUpdateTransaction(t *testing.T) {
	ts, log := newRecordingServer(t)
	defer ts.Close()
	c := clientFor(ts)

	body := map[string]interface{}{"transaction": map[string]interface{}{"memo": "test"}}
	_, err := c.UpdateTransaction("plan-1", "tx-1", body)
	if err != nil {
		t.Fatal(err)
	}
	if log.Method != "PUT" {
		t.Errorf("method = %s, want PUT", log.Method)
	}
	if log.Path != "/budgets/plan-1/transactions/tx-1" {
		t.Errorf("path = %s", log.Path)
	}
	if log.Body == "" {
		t.Error("expected non-empty body")
	}
}

func TestDeleteTransaction(t *testing.T) {
	ts, log := newRecordingServer(t)
	defer ts.Close()
	c := clientFor(ts)

	_, err := c.DeleteTransaction("plan-1", "tx-1")
	if err != nil {
		t.Fatal(err)
	}
	if log.Method != "DELETE" {
		t.Errorf("method = %s, want DELETE", log.Method)
	}
	if log.Path != "/budgets/plan-1/transactions/tx-1" {
		t.Errorf("path = %s", log.Path)
	}
	if log.Body != "" {
		t.Errorf("expected no body for DELETE, got %q", log.Body)
	}
}

func TestImportTransactions(t *testing.T) {
	ts, log := newRecordingServer(t)
	defer ts.Close()
	c := clientFor(ts)

	_, err := c.ImportTransactions("plan-1")
	if err != nil {
		t.Fatal(err)
	}
	if log.Method != "POST" {
		t.Errorf("method = %s, want POST", log.Method)
	}
	if log.Path != "/budgets/plan-1/transactions/import" {
		t.Errorf("path = %s", log.Path)
	}
}

func TestGetTransactionsByAccount(t *testing.T) {
	t.Run("no_filters", func(t *testing.T) {
		ts, log := newRecordingServer(t)
		defer ts.Close()
		c := clientFor(ts)

		_, err := c.GetTransactionsByAccount("plan-1", "acc-1", "", "", 0)
		if err != nil {
			t.Fatal(err)
		}
		if log.Path != "/budgets/plan-1/accounts/acc-1/transactions" {
			t.Errorf("path = %s", log.Path)
		}
		if log.Query != "" {
			t.Errorf("query = %s, want empty", log.Query)
		}
	})

	t.Run("with_filters", func(t *testing.T) {
		ts, log := newRecordingServer(t)
		defer ts.Close()
		c := clientFor(ts)

		_, err := c.GetTransactionsByAccount("plan-1", "acc-1", "2024-01-01", "unapproved", 10)
		if err != nil {
			t.Fatal(err)
		}
		if log.Path != "/budgets/plan-1/accounts/acc-1/transactions" {
			t.Errorf("path = %s", log.Path)
		}
		if log.Query != "last_knowledge_of_server=10&since_date=2024-01-01&type=unapproved" {
			t.Errorf("query = %s", log.Query)
		}
	})
}

func TestGetTransactionsByCategory(t *testing.T) {
	t.Run("no_filters", func(t *testing.T) {
		ts, log := newRecordingServer(t)
		defer ts.Close()
		c := clientFor(ts)

		_, err := c.GetTransactionsByCategory("plan-1", "cat-1", "", "", 0)
		if err != nil {
			t.Fatal(err)
		}
		if log.Path != "/budgets/plan-1/categories/cat-1/transactions" {
			t.Errorf("path = %s", log.Path)
		}
		if log.Query != "" {
			t.Errorf("query = %s, want empty", log.Query)
		}
	})

	t.Run("with_since_date", func(t *testing.T) {
		ts, log := newRecordingServer(t)
		defer ts.Close()
		c := clientFor(ts)

		_, err := c.GetTransactionsByCategory("plan-1", "cat-1", "2024-05-01", "", 0)
		if err != nil {
			t.Fatal(err)
		}
		if log.Query != "since_date=2024-05-01" {
			t.Errorf("query = %s, want since_date=2024-05-01", log.Query)
		}
	})
}

func TestGetTransactionsByPayee(t *testing.T) {
	t.Run("no_filters", func(t *testing.T) {
		ts, log := newRecordingServer(t)
		defer ts.Close()
		c := clientFor(ts)

		_, err := c.GetTransactionsByPayee("plan-1", "payee-1", "", "", 0)
		if err != nil {
			t.Fatal(err)
		}
		if log.Path != "/budgets/plan-1/payees/payee-1/transactions" {
			t.Errorf("path = %s", log.Path)
		}
		if log.Query != "" {
			t.Errorf("query = %s, want empty", log.Query)
		}
	})

	t.Run("with_type", func(t *testing.T) {
		ts, log := newRecordingServer(t)
		defer ts.Close()
		c := clientFor(ts)

		_, err := c.GetTransactionsByPayee("plan-1", "payee-1", "", "uncategorized", 0)
		if err != nil {
			t.Fatal(err)
		}
		if log.Query != "type=uncategorized" {
			t.Errorf("query = %s, want type=uncategorized", log.Query)
		}
	})
}

func TestGetTransactionsByMonth(t *testing.T) {
	t.Run("no_filters", func(t *testing.T) {
		ts, log := newRecordingServer(t)
		defer ts.Close()
		c := clientFor(ts)

		_, err := c.GetTransactionsByMonth("plan-1", "2024-03-01", "", "", 0)
		if err != nil {
			t.Fatal(err)
		}
		if log.Path != "/budgets/plan-1/months/2024-03-01/transactions" {
			t.Errorf("path = %s", log.Path)
		}
		if log.Query != "" {
			t.Errorf("query = %s, want empty", log.Query)
		}
	})

	t.Run("with_all_filters", func(t *testing.T) {
		ts, log := newRecordingServer(t)
		defer ts.Close()
		c := clientFor(ts)

		_, err := c.GetTransactionsByMonth("plan-1", "2024-03-01", "2024-03-15", "unapproved", 50)
		if err != nil {
			t.Fatal(err)
		}
		if log.Query != "last_knowledge_of_server=50&since_date=2024-03-15&type=unapproved" {
			t.Errorf("query = %s", log.Query)
		}
	})
}

// ---------------------------------------------------------------------------
// Payees
// ---------------------------------------------------------------------------

func TestGetPayees(t *testing.T) {
	t.Run("no_knowledge", func(t *testing.T) {
		ts, log := newRecordingServer(t)
		defer ts.Close()
		c := clientFor(ts)

		_, err := c.GetPayees("plan-1", 0)
		if err != nil {
			t.Fatal(err)
		}
		if log.Path != "/budgets/plan-1/payees" {
			t.Errorf("path = %s", log.Path)
		}
		if log.Query != "" {
			t.Errorf("query = %s, want empty", log.Query)
		}
	})

	t.Run("with_knowledge", func(t *testing.T) {
		ts, log := newRecordingServer(t)
		defer ts.Close()
		c := clientFor(ts)

		_, err := c.GetPayees("plan-1", 50)
		if err != nil {
			t.Fatal(err)
		}
		if log.Path != "/budgets/plan-1/payees" {
			t.Errorf("path = %s", log.Path)
		}
		if log.Query != "last_knowledge_of_server=50" {
			t.Errorf("query = %s", log.Query)
		}
	})
}

func TestGetPayeeByID(t *testing.T) {
	ts, log := newRecordingServer(t)
	defer ts.Close()
	c := clientFor(ts)

	_, err := c.GetPayeeByID("plan-1", "payee-1")
	if err != nil {
		t.Fatal(err)
	}
	if log.Method != "GET" {
		t.Errorf("method = %s, want GET", log.Method)
	}
	if log.Path != "/budgets/plan-1/payees/payee-1" {
		t.Errorf("path = %s, want /budgets/plan-1/payees/payee-1", log.Path)
	}
}

func TestCreatePayee(t *testing.T) {
	ts, log := newRecordingServer(t)
	defer ts.Close()
	c := clientFor(ts)

	body := map[string]interface{}{"payee": map[string]interface{}{"name": "Test"}}
	_, err := c.CreatePayee("plan-1", body)
	if err != nil {
		t.Fatal(err)
	}
	if log.Method != "POST" {
		t.Errorf("method = %s, want POST", log.Method)
	}
	if log.Path != "/budgets/plan-1/payees" {
		t.Errorf("path = %s, want /budgets/plan-1/payees", log.Path)
	}
	if log.Body == "" {
		t.Error("expected non-empty body")
	}
}

func TestUpdatePayee(t *testing.T) {
	ts, log := newRecordingServer(t)
	defer ts.Close()
	c := clientFor(ts)

	body := map[string]interface{}{"payee": map[string]interface{}{"name": "Updated"}}
	_, err := c.UpdatePayee("plan-1", "payee-1", body)
	if err != nil {
		t.Fatal(err)
	}
	if log.Method != "PATCH" {
		t.Errorf("method = %s, want PATCH", log.Method)
	}
	if log.Path != "/budgets/plan-1/payees/payee-1" {
		t.Errorf("path = %s", log.Path)
	}
	if log.Body == "" {
		t.Error("expected non-empty body")
	}
}

func TestPayeeLocations(t *testing.T) {
	t.Run("list", func(t *testing.T) {
		ts, log := newRecordingServer(t)
		defer ts.Close()
		c := clientFor(ts)

		_, err := c.GetPayeeLocations("plan-1")
		if err != nil {
			t.Fatal(err)
		}
		if log.Method != "GET" {
			t.Errorf("method = %s, want GET", log.Method)
		}
		if log.Path != "/budgets/plan-1/payee_locations" {
			t.Errorf("path = %s", log.Path)
		}
	})

	t.Run("get_by_id", func(t *testing.T) {
		ts, log := newRecordingServer(t)
		defer ts.Close()
		c := clientFor(ts)

		_, err := c.GetPayeeLocationByID("plan-1", "loc-1")
		if err != nil {
			t.Fatal(err)
		}
		if log.Path != "/budgets/plan-1/payee_locations/loc-1" {
			t.Errorf("path = %s", log.Path)
		}
	})

	t.Run("list_by_payee", func(t *testing.T) {
		ts, log := newRecordingServer(t)
		defer ts.Close()
		c := clientFor(ts)

		_, err := c.GetPayeeLocationsByPayee("plan-1", "payee-1")
		if err != nil {
			t.Fatal(err)
		}
		if log.Path != "/budgets/plan-1/payees/payee-1/payee_locations" {
			t.Errorf("path = %s", log.Path)
		}
	})
}

// ---------------------------------------------------------------------------
// Scheduled Transactions
// ---------------------------------------------------------------------------

func TestScheduledTransactions(t *testing.T) {
	t.Run("list_no_knowledge", func(t *testing.T) {
		ts, log := newRecordingServer(t)
		defer ts.Close()
		c := clientFor(ts)

		_, err := c.GetScheduledTransactions("plan-1", 0)
		if err != nil {
			t.Fatal(err)
		}
		if log.Path != "/budgets/plan-1/scheduled_transactions" {
			t.Errorf("path = %s", log.Path)
		}
		if log.Query != "" {
			t.Errorf("query = %s, want empty", log.Query)
		}
	})

	t.Run("list_with_knowledge", func(t *testing.T) {
		ts, log := newRecordingServer(t)
		defer ts.Close()
		c := clientFor(ts)

		_, err := c.GetScheduledTransactions("plan-1", 200)
		if err != nil {
			t.Fatal(err)
		}
		if log.Path != "/budgets/plan-1/scheduled_transactions" {
			t.Errorf("path = %s", log.Path)
		}
		if log.Query != "last_knowledge_of_server=200" {
			t.Errorf("query = %s", log.Query)
		}
	})

	t.Run("get", func(t *testing.T) {
		ts, log := newRecordingServer(t)
		defer ts.Close()
		c := clientFor(ts)

		_, err := c.GetScheduledTransactionByID("plan-1", "st-1")
		if err != nil {
			t.Fatal(err)
		}
		if log.Method != "GET" {
			t.Errorf("method = %s, want GET", log.Method)
		}
		if log.Path != "/budgets/plan-1/scheduled_transactions/st-1" {
			t.Errorf("path = %s", log.Path)
		}
	})

	t.Run("create", func(t *testing.T) {
		ts, log := newRecordingServer(t)
		defer ts.Close()
		c := clientFor(ts)

		body := map[string]interface{}{"scheduled_transaction": map[string]interface{}{
			"account_id": "acc-1",
			"date":       "2024-04-01",
			"amount":     -100000,
			"frequency":  "monthly",
		}}
		_, err := c.CreateScheduledTransaction("plan-1", body)
		if err != nil {
			t.Fatal(err)
		}
		if log.Method != "POST" {
			t.Errorf("method = %s, want POST", log.Method)
		}
		if log.Path != "/budgets/plan-1/scheduled_transactions" {
			t.Errorf("path = %s", log.Path)
		}
		if log.Body == "" {
			t.Error("expected non-empty body")
		}
		// Verify the body uses "date" not "date_first"
		var parsed map[string]interface{}
		if err := json.Unmarshal([]byte(log.Body), &parsed); err != nil {
			t.Fatalf("failed to parse body: %v", err)
		}
		st, ok := parsed["scheduled_transaction"].(map[string]interface{})
		if !ok {
			t.Fatal("missing scheduled_transaction in body")
		}
		if _, exists := st["date_first"]; exists {
			t.Error("body contains 'date_first' — API expects 'date' for create")
		}
		if st["date"] != "2024-04-01" {
			t.Errorf("date = %v, want 2024-04-01", st["date"])
		}
	})

	t.Run("update", func(t *testing.T) {
		ts, log := newRecordingServer(t)
		defer ts.Close()
		c := clientFor(ts)

		body := map[string]interface{}{"scheduled_transaction": map[string]interface{}{"memo": "test"}}
		_, err := c.UpdateScheduledTransaction("plan-1", "st-1", body)
		if err != nil {
			t.Fatal(err)
		}
		if log.Method != "PUT" {
			t.Errorf("method = %s, want PUT", log.Method)
		}
		if log.Path != "/budgets/plan-1/scheduled_transactions/st-1" {
			t.Errorf("path = %s", log.Path)
		}
		if log.Body == "" {
			t.Error("expected non-empty body")
		}
	})

	t.Run("delete", func(t *testing.T) {
		ts, log := newRecordingServer(t)
		defer ts.Close()
		c := clientFor(ts)

		_, err := c.DeleteScheduledTransaction("plan-1", "st-1")
		if err != nil {
			t.Fatal(err)
		}
		if log.Method != "DELETE" {
			t.Errorf("method = %s, want DELETE", log.Method)
		}
		if log.Path != "/budgets/plan-1/scheduled_transactions/st-1" {
			t.Errorf("path = %s", log.Path)
		}
		if log.Body != "" {
			t.Errorf("expected no body for DELETE, got %q", log.Body)
		}
	})
}

// ---------------------------------------------------------------------------
// Months
// ---------------------------------------------------------------------------

func TestMonths(t *testing.T) {
	t.Run("list_no_knowledge", func(t *testing.T) {
		ts, log := newRecordingServer(t)
		defer ts.Close()
		c := clientFor(ts)

		_, err := c.GetMonths("plan-1", 0)
		if err != nil {
			t.Fatal(err)
		}
		if log.Path != "/budgets/plan-1/months" {
			t.Errorf("path = %s", log.Path)
		}
		if log.Query != "" {
			t.Errorf("query = %s, want empty", log.Query)
		}
	})

	t.Run("list_with_knowledge", func(t *testing.T) {
		ts, log := newRecordingServer(t)
		defer ts.Close()
		c := clientFor(ts)

		_, err := c.GetMonths("plan-1", 25)
		if err != nil {
			t.Fatal(err)
		}
		if log.Path != "/budgets/plan-1/months" {
			t.Errorf("path = %s", log.Path)
		}
		if log.Query != "last_knowledge_of_server=25" {
			t.Errorf("query = %s, want last_knowledge_of_server=25", log.Query)
		}
	})

	t.Run("get", func(t *testing.T) {
		ts, log := newRecordingServer(t)
		defer ts.Close()
		c := clientFor(ts)

		_, err := c.GetMonth("plan-1", "2024-03-01")
		if err != nil {
			t.Fatal(err)
		}
		if log.Method != "GET" {
			t.Errorf("method = %s, want GET", log.Method)
		}
		if log.Path != "/budgets/plan-1/months/2024-03-01" {
			t.Errorf("path = %s", log.Path)
		}
	})
}

// ---------------------------------------------------------------------------
// Money Movements
// ---------------------------------------------------------------------------

func TestMoneyMovements(t *testing.T) {
	t.Run("list", func(t *testing.T) {
		ts, log := newRecordingServer(t)
		defer ts.Close()
		c := clientFor(ts)

		_, err := c.GetMoneyMovements("plan-1")
		if err != nil {
			t.Fatal(err)
		}
		if log.Method != "GET" {
			t.Errorf("method = %s, want GET", log.Method)
		}
		if log.Path != "/budgets/plan-1/money_movements" {
			t.Errorf("path = %s", log.Path)
		}
	})

	t.Run("list_by_month", func(t *testing.T) {
		ts, log := newRecordingServer(t)
		defer ts.Close()
		c := clientFor(ts)

		_, err := c.GetMoneyMovementsByMonth("plan-1", "2024-03-01")
		if err != nil {
			t.Fatal(err)
		}
		if log.Path != "/budgets/plan-1/months/2024-03-01/money_movements" {
			t.Errorf("path = %s", log.Path)
		}
	})

	t.Run("groups_list", func(t *testing.T) {
		ts, log := newRecordingServer(t)
		defer ts.Close()
		c := clientFor(ts)

		_, err := c.GetMoneyMovementGroups("plan-1")
		if err != nil {
			t.Fatal(err)
		}
		if log.Method != "GET" {
			t.Errorf("method = %s, want GET", log.Method)
		}
		if log.Path != "/budgets/plan-1/money_movement_groups" {
			t.Errorf("path = %s", log.Path)
		}
	})

	t.Run("groups_list_by_month", func(t *testing.T) {
		ts, log := newRecordingServer(t)
		defer ts.Close()
		c := clientFor(ts)

		_, err := c.GetMoneyMovementGroupsByMonth("plan-1", "2024-03-01")
		if err != nil {
			t.Fatal(err)
		}
		if log.Path != "/budgets/plan-1/months/2024-03-01/money_movement_groups" {
			t.Errorf("path = %s", log.Path)
		}
	})
}
