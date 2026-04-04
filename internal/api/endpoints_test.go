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
}

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

func TestGetAccounts(t *testing.T) {
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

func TestGetCategories(t *testing.T) {
	ts, log := newRecordingServer(t)
	defer ts.Close()
	c := clientFor(ts)

	_, err := c.GetCategories("plan-1", 0)
	if err != nil {
		t.Fatal(err)
	}
	if log.Path != "/budgets/plan-1/categories" {
		t.Errorf("path = %s", log.Path)
	}
	if log.Query != "" {
		t.Errorf("query = %s, want empty", log.Query)
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
		t.Errorf("path = %s", log.Path)
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
	if log.Path != "/budgets/plan-1/months/2024-03-01/categories/cat-1" {
		t.Errorf("path = %s", log.Path)
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
		t.Errorf("path = %s", log.Path)
	}
}

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
			if log.Query != tt.wantQuery {
				t.Errorf("query = %s, want %s", log.Query, tt.wantQuery)
			}
		})
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
		t.Errorf("method = %s", log.Method)
	}
	if log.Path != "/budgets/plan-1/transactions" {
		t.Errorf("path = %s", log.Path)
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
		t.Errorf("method = %s", log.Method)
	}
	if log.Path != "/budgets/plan-1/transactions/import" {
		t.Errorf("path = %s", log.Path)
	}
}

func TestGetTransactionsByAccount(t *testing.T) {
	ts, log := newRecordingServer(t)
	defer ts.Close()
	c := clientFor(ts)

	_, err := c.GetTransactionsByAccount("plan-1", "acc-1", "2024-01-01", "", 0)
	if err != nil {
		t.Fatal(err)
	}
	if log.Path != "/budgets/plan-1/accounts/acc-1/transactions" {
		t.Errorf("path = %s", log.Path)
	}
}

func TestGetTransactionsByCategory(t *testing.T) {
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
}

func TestGetTransactionsByPayee(t *testing.T) {
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
}

func TestGetTransactionsByMonth(t *testing.T) {
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
}

func TestGetPayees(t *testing.T) {
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
		t.Errorf("method = %s", log.Method)
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
		t.Errorf("method = %s", log.Method)
	}
	if log.Path != "/budgets/plan-1/payees/payee-1" {
		t.Errorf("path = %s", log.Path)
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

func TestScheduledTransactions(t *testing.T) {
	t.Run("list", func(t *testing.T) {
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
		if log.Path != "/budgets/plan-1/scheduled_transactions/st-1" {
			t.Errorf("path = %s", log.Path)
		}
	})

	t.Run("create", func(t *testing.T) {
		ts, log := newRecordingServer(t)
		defer ts.Close()
		c := clientFor(ts)

		body := map[string]interface{}{"scheduled_transaction": map[string]interface{}{"amount": -100000}}
		_, err := c.CreateScheduledTransaction("plan-1", body)
		if err != nil {
			t.Fatal(err)
		}
		if log.Method != "POST" {
			t.Errorf("method = %s", log.Method)
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
			t.Errorf("method = %s", log.Method)
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
			t.Errorf("method = %s", log.Method)
		}
		if log.Path != "/budgets/plan-1/scheduled_transactions/st-1" {
			t.Errorf("path = %s", log.Path)
		}
	})
}

func TestMonths(t *testing.T) {
	t.Run("list", func(t *testing.T) {
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
	})

	t.Run("get", func(t *testing.T) {
		ts, log := newRecordingServer(t)
		defer ts.Close()
		c := clientFor(ts)

		_, err := c.GetMonth("plan-1", "2024-03-01")
		if err != nil {
			t.Fatal(err)
		}
		if log.Path != "/budgets/plan-1/months/2024-03-01" {
			t.Errorf("path = %s", log.Path)
		}
	})
}

func TestMoneyMovements(t *testing.T) {
	t.Run("list", func(t *testing.T) {
		ts, log := newRecordingServer(t)
		defer ts.Close()
		c := clientFor(ts)

		_, err := c.GetMoneyMovements("plan-1")
		if err != nil {
			t.Fatal(err)
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
