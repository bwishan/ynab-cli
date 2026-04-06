package mcp

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestServer() *Server {
	reg := NewRegistry(newTestRoot())
	runner := func(args []string) (string, error) {
		return `{"data":{}}`, nil
	}
	handlers := NewToolHandlers(reg, runner)
	return NewServer(reg, handlers, "test")
}

func TestDispatchInitialize(t *testing.T) {
	srv := newTestServer()
	req := &Request{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`1`),
		Method:  "initialize",
	}

	resp := srv.dispatch(req)
	if resp == nil {
		t.Fatal("expected response")
	}
	if resp.Error != nil {
		t.Fatalf("unexpected error: %+v", resp.Error)
	}

	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatal("expected map result")
	}
	serverInfo := result["serverInfo"].(map[string]any)
	if serverInfo["name"] != "ynab-cli" {
		t.Errorf("expected server name 'ynab-cli', got %q", serverInfo["name"])
	}
}

func TestDispatchToolsList(t *testing.T) {
	srv := newTestServer()
	req := &Request{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`2`),
		Method:  "tools/list",
	}

	resp := srv.dispatch(req)
	result := resp.Result.(map[string]any)
	tools := result["tools"].([]ToolDef)
	if len(tools) != 2 {
		t.Fatalf("expected 2 tools, got %d", len(tools))
	}
}

func TestDispatchToolsCall(t *testing.T) {
	srv := newTestServer()
	params, _ := json.Marshal(ToolCallParams{
		Name:      "search_commands",
		Arguments: json.RawMessage(`{"resource":"plans"}`),
	})
	req := &Request{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`3`),
		Method:  "tools/call",
		Params:  params,
	}

	resp := srv.dispatch(req)
	if resp.Error != nil {
		t.Fatalf("unexpected error: %+v", resp.Error)
	}
}

func TestDispatchNotification(t *testing.T) {
	srv := newTestServer()
	req := &Request{
		JSONRPC: "2.0",
		Method:  "notifications/initialized",
	}

	resp := srv.dispatch(req)
	if resp != nil {
		t.Error("expected nil response for notification")
	}
}

func TestDispatchUnknownMethod(t *testing.T) {
	srv := newTestServer()
	req := &Request{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`4`),
		Method:  "unknown/method",
	}

	resp := srv.dispatch(req)
	if resp.Error == nil {
		t.Fatal("expected error for unknown method")
	}
	if resp.Error.Code != -32601 {
		t.Errorf("expected code -32601, got %d", resp.Error.Code)
	}
}

func TestHTTPHandler(t *testing.T) {
	srv := newTestServer()

	body, _ := json.Marshal(Request{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`1`),
		Method:  "initialize",
	})

	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.handleHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var rpcResp Response
	respBody, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(respBody, &rpcResp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if rpcResp.Error != nil {
		t.Fatalf("unexpected error: %+v", rpcResp.Error)
	}
}

func TestHTTPMethodNotAllowed(t *testing.T) {
	srv := newTestServer()
	req := httptest.NewRequest(http.MethodGet, "/mcp", nil)
	w := httptest.NewRecorder()

	srv.handleHTTP(w, req)

	if w.Result().StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Result().StatusCode)
	}
}

func TestStdioRoundTrip(t *testing.T) {
	srv := newTestServer()

	// Build a request
	reqJSON, _ := json.Marshal(Request{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`1`),
		Method:  "tools/list",
	})

	// We can't easily test the full Run loop (it reads from os.Stdin),
	// but we can test dispatch directly for each message type.
	var req Request
	json.Unmarshal(reqJSON, &req)

	resp := srv.dispatch(&req)
	if resp == nil {
		t.Fatal("expected response")
	}

	// Verify it serializes correctly
	out, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal response: %v", err)
	}
	if !strings.Contains(string(out), "search_commands") {
		t.Error("response should contain tool definitions")
	}
}
