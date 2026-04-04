// Package api provides an HTTP client for the YNAB API v1 using only Go's standard library.
package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	BaseURL    = "https://api.ynab.com/v1"
	userAgent  = "ynab-cli"
)

// Client is the YNAB API client. It handles authentication, rate limiting,
// and response parsing using only Go's standard library.
type Client struct {
	token      string
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new YNAB API client with the given access token.
func NewClient(token string) *Client {
	return &Client{
		token:   token,
		baseURL: BaseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ErrorResponse represents a YNAB API error.
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains the error information.
type ErrorDetail struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Detail string `json:"detail"`
}

// APIError is returned when the YNAB API returns a non-2xx status.
type APIError struct {
	StatusCode int
	ErrorDetail
}

func (e *APIError) Error() string {
	return fmt.Sprintf("ynab api error %d (%s): %s", e.StatusCode, e.Name, e.Detail)
}

// doRequest executes an HTTP request and returns the raw response body.
func (c *Client) doRequest(method, path string, query url.Values, body interface{}) ([]byte, error) {
	u := c.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, u, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("User-Agent", userAgent)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errResp ErrorResponse
		if jsonErr := json.Unmarshal(data, &errResp); jsonErr == nil && errResp.Error.ID != "" {
			return nil, &APIError{
				StatusCode:  resp.StatusCode,
				ErrorDetail: errResp.Error,
			}
		}
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			ErrorDetail: ErrorDetail{
				ID:     fmt.Sprintf("%d", resp.StatusCode),
				Name:   http.StatusText(resp.StatusCode),
				Detail: string(data),
			},
		}
	}

	return data, nil
}

// Get performs an HTTP GET request.
func (c *Client) Get(path string, query url.Values) ([]byte, error) {
	return c.doRequest(http.MethodGet, path, query, nil)
}

// Post performs an HTTP POST request.
func (c *Client) Post(path string, body interface{}) ([]byte, error) {
	return c.doRequest(http.MethodPost, path, nil, body)
}

// Put performs an HTTP PUT request.
func (c *Client) Put(path string, body interface{}) ([]byte, error) {
	return c.doRequest(http.MethodPut, path, nil, body)
}

// Patch performs an HTTP PATCH request.
func (c *Client) Patch(path string, body interface{}) ([]byte, error) {
	return c.doRequest(http.MethodPatch, path, nil, body)
}

// Delete performs an HTTP DELETE request.
func (c *Client) Delete(path string) ([]byte, error) {
	return c.doRequest(http.MethodDelete, path, nil, nil)
}
