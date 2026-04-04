package api

import (
	"fmt"
	"net/url"
	"strconv"
)

// GetAccounts returns all accounts for a plan.
// GET /plans/{plan_id}/accounts?last_knowledge_of_server=N
func (c *Client) GetAccounts(planID string, lastKnowledge int64) ([]byte, error) {
	q := url.Values{}
	if lastKnowledge > 0 {
		q.Set("last_knowledge_of_server", strconv.FormatInt(lastKnowledge, 10))
	}
	return c.Get(fmt.Sprintf("/budgets/%s/accounts", planID), q)
}

// GetAccountByID returns a single account by ID.
// GET /plans/{plan_id}/accounts/{account_id}
func (c *Client) GetAccountByID(planID, accountID string) ([]byte, error) {
	return c.Get(fmt.Sprintf("/budgets/%s/accounts/%s", planID, accountID), nil)
}

// CreateAccount creates a new account in a plan.
// POST /plans/{plan_id}/accounts
func (c *Client) CreateAccount(planID string, body interface{}) ([]byte, error) {
	return c.Post(fmt.Sprintf("/budgets/%s/accounts", planID), body)
}
