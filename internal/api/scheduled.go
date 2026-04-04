package api

import (
	"fmt"
	"net/url"
	"strconv"
)

// GetScheduledTransactions returns all scheduled transactions for a plan.
// GET /plans/{plan_id}/scheduled_transactions?last_knowledge_of_server=N
func (c *Client) GetScheduledTransactions(planID string, lastKnowledge int64) ([]byte, error) {
	q := url.Values{}
	if lastKnowledge > 0 {
		q.Set("last_knowledge_of_server", strconv.FormatInt(lastKnowledge, 10))
	}
	return c.Get(fmt.Sprintf("/budgets/%s/scheduled_transactions", planID), q)
}

// GetScheduledTransactionByID returns a single scheduled transaction by ID.
// GET /plans/{plan_id}/scheduled_transactions/{id}
func (c *Client) GetScheduledTransactionByID(planID, transactionID string) ([]byte, error) {
	return c.Get(fmt.Sprintf("/budgets/%s/scheduled_transactions/%s", planID, transactionID), nil)
}

// CreateScheduledTransaction creates a new scheduled transaction.
// POST /plans/{plan_id}/scheduled_transactions  body: { scheduled_transaction: {...} }
func (c *Client) CreateScheduledTransaction(planID string, body interface{}) ([]byte, error) {
	return c.Post(fmt.Sprintf("/budgets/%s/scheduled_transactions", planID), body)
}

// UpdateScheduledTransaction updates an existing scheduled transaction.
// PUT /plans/{plan_id}/scheduled_transactions/{id}  body: { scheduled_transaction: {...} }
func (c *Client) UpdateScheduledTransaction(planID, transactionID string, body interface{}) ([]byte, error) {
	return c.Put(fmt.Sprintf("/budgets/%s/scheduled_transactions/%s", planID, transactionID), body)
}

// DeleteScheduledTransaction deletes a scheduled transaction.
// DELETE /plans/{plan_id}/scheduled_transactions/{id}
func (c *Client) DeleteScheduledTransaction(planID, transactionID string) ([]byte, error) {
	return c.Delete(fmt.Sprintf("/budgets/%s/scheduled_transactions/%s", planID, transactionID))
}
