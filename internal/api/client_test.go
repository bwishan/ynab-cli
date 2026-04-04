package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// newTestClient creates a Client pointed at the given test server.
func newTestClient(ts *httptest.Server) *Client {
	c := NewClient("test-token-123")
	c.baseURL = ts.URL
	c.httpClient = ts.Client()
	return c
}

func TestAuthorizationHeader(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := r.Header.Get("Authorization")
		if got != "Bearer test-token-123" {
			t.Errorf("Authorization = %q, want %q", got, "Bearer test-token-123")
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, _ = c.Get("/test", nil)
}

func TestAcceptHeader(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := r.Header.Get("Accept")
		if got != "application/json" {
			t.Errorf("Accept = %q, want %q", got, "application/json")
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, _ = c.Get("/test", nil)
}

func TestUserAgentHeader(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := r.Header.Get("User-Agent")
		if got != "ynab-cli" {
			t.Errorf("User-Agent = %q, want %q", got, "ynab-cli")
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, _ = c.Get("/test", nil)
}

func TestContentTypeHeaderOnPost(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ct := r.Header.Get("Content-Type")
		if ct != "application/json" {
			t.Errorf("Content-Type = %q, want %q", ct, "application/json")
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, _ = c.Post("/test", map[string]string{"key": "val"})
}

func TestNoContentTypeHeaderOnGet(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ct := r.Header.Get("Content-Type")
		if ct != "" {
			t.Errorf("Content-Type on GET = %q, want empty", ct)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, _ = c.Get("/test", nil)
}

func TestHTTPMethods(t *testing.T) {
	tests := []struct {
		name   string
		call   func(c *Client) ([]byte, error)
		method string
	}{
		{"GET", func(c *Client) ([]byte, error) { return c.Get("/x", nil) }, "GET"},
		{"POST", func(c *Client) ([]byte, error) { return c.Post("/x", map[string]string{"a": "b"}) }, "POST"},
		{"PUT", func(c *Client) ([]byte, error) { return c.Put("/x", map[string]string{"a": "b"}) }, "PUT"},
		{"PATCH", func(c *Client) ([]byte, error) { return c.Patch("/x", map[string]string{"a": "b"}) }, "PATCH"},
		{"DELETE", func(c *Client) ([]byte, error) { return c.Delete("/x") }, "DELETE"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != tt.method {
					t.Errorf("method = %q, want %q", r.Method, tt.method)
				}
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{}`))
			}))
			defer ts.Close()
			c := newTestClient(ts)
			_, _ = tt.call(c)
		})
	}
}

func TestQueryParametersEncoded(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("since_date"); got != "2024-01-01" {
			t.Errorf("since_date = %q, want %q", got, "2024-01-01")
		}
		if got := r.URL.Query().Get("type"); got != "unapproved" {
			t.Errorf("type = %q, want %q", got, "unapproved")
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer ts.Close()

	c := newTestClient(ts)
	q := make(map[string][]string)
	q["since_date"] = []string{"2024-01-01"}
	q["type"] = []string{"unapproved"}
	_, _ = c.Get("/test", q)
}

func TestRequestBodyJSONMarshal(t *testing.T) {
	type payload struct {
		Name   string `json:"name"`
		Amount int    `json:"amount"`
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var p payload
		if err := json.Unmarshal(body, &p); err != nil {
			t.Fatalf("failed to unmarshal body: %v", err)
		}
		if p.Name != "Groceries" || p.Amount != 5000 {
			t.Errorf("body = %+v, want {Name:Groceries Amount:5000}", p)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, _ = c.Post("/test", payload{Name: "Groceries", Amount: 5000})
}

func TestSuccessfulResponseReturnsBody(t *testing.T) {
	want := `{"data":{"user":{"id":"abc"}}}`
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(want))
	}))
	defer ts.Close()

	c := newTestClient(ts)
	got, err := c.Get("/test", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != want {
		t.Errorf("body = %q, want %q", string(got), want)
	}
}

func TestErrorResponseParsedIntoAPIError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":{"id":"404","name":"not_found","detail":"Budget not found"}}`))
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.Get("/test", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != 404 {
		t.Errorf("StatusCode = %d, want 404", apiErr.StatusCode)
	}
	if apiErr.ID != "404" {
		t.Errorf("ID = %q, want %q", apiErr.ID, "404")
	}
	if apiErr.Name != "not_found" {
		t.Errorf("Name = %q, want %q", apiErr.Name, "not_found")
	}
	if apiErr.Detail != "Budget not found" {
		t.Errorf("Detail = %q, want %q", apiErr.Detail, "Budget not found")
	}
}

func TestErrorResponseErrorString(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":{"id":"404","name":"not_found","detail":"Budget not found"}}`))
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.Get("/test", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	want := "ynab api error 404 (not_found): Budget not found"
	if err.Error() != want {
		t.Errorf("error string = %q, want %q", err.Error(), want)
	}
}

func TestNonJSONErrorResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("Bad Gateway"))
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.Get("/test", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != 502 {
		t.Errorf("StatusCode = %d, want 502", apiErr.StatusCode)
	}
	if apiErr.Name != "Bad Gateway" {
		t.Errorf("Name = %q, want %q", apiErr.Name, "Bad Gateway")
	}
	if apiErr.Detail != "Bad Gateway" {
		t.Errorf("Detail = %q, want %q", apiErr.Detail, "Bad Gateway")
	}
}

func TestRateLimitError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"error":{"id":"429","name":"too_many_requests","detail":"Rate limit exceeded. Retry after 2 seconds."}}`))
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.Get("/test", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != 429 {
		t.Errorf("StatusCode = %d, want 429", apiErr.StatusCode)
	}
	if apiErr.Name != "too_many_requests" {
		t.Errorf("Name = %q, want %q", apiErr.Name, "too_many_requests")
	}
}

func TestDeleteSendsNoBody(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if len(body) != 0 {
			t.Errorf("DELETE body = %q, want empty", string(body))
		}
		if ct := r.Header.Get("Content-Type"); ct != "" {
			t.Errorf("DELETE Content-Type = %q, want empty", ct)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, _ = c.Delete("/test")
}

func TestEmptyJSONErrorFallback(t *testing.T) {
	// Error JSON without an id field should use fallback path.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":{}}`))
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.Get("/test", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	// Since error.id is empty, it should fall through to the fallback path
	if apiErr.StatusCode != 500 {
		t.Errorf("StatusCode = %d, want 500", apiErr.StatusCode)
	}
	if apiErr.Name != "Internal Server Error" {
		t.Errorf("Name = %q, want %q", apiErr.Name, "Internal Server Error")
	}
}

func TestPostWithNilBodySendsNoContent(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if len(body) != 0 {
			t.Errorf("POST with nil body should have empty body, got %q", string(body))
		}
		if ct := r.Header.Get("Content-Type"); ct != "" {
			t.Errorf("POST with nil body should have no Content-Type, got %q", ct)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, _ = c.Post("/test", nil)
}

func TestMarshalErrorReturned(t *testing.T) {
	c := NewClient("tok")
	// channels cannot be JSON-marshaled
	_, err := c.Post("/test", make(chan int))
	if err == nil {
		t.Fatal("expected marshal error, got nil")
	}
	got := err.Error()
	if got == "" {
		t.Error("expected non-empty error message")
	}
}

func TestNoQueryParamsOmitsQuestionMark(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "" {
			t.Errorf("expected no query string, got %q", r.URL.RawQuery)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, _ = c.Get("/test", nil)
}

func TestStatus201IsSuccess(t *testing.T) {
	want := `{"data":{"id":"new"}}`
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(want))
	}))
	defer ts.Close()

	c := newTestClient(ts)
	got, err := c.Post("/test", map[string]string{"name": "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != want {
		t.Errorf("body = %q, want %q", string(got), want)
	}
}
