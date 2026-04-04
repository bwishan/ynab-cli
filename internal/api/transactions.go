package api

import (
	"fmt"
	"net/url"
	"strconv"
)

// buildTransactionQuery builds common query parameters for transaction list endpoints.
func buildTransactionQuery(sinceDate, txType string, lastKnowledge int64) url.Values {
	q := url.Values{}
	if sinceDate != "" {
		q.Set("since_date", sinceDate)
	}
	if txType != "" {
		q.Set("type", txType)
	}
	if lastKnowledge > 0 {
		q.Set("last_knowledge_of_server", strconv.FormatInt(lastKnowledge, 10))
	}
	return q
}

// GetTransactions returns all transactions for a plan.
// GET /budgets/{plan_id}/transactions
func (c *Client) GetTransactions(planID string, sinceDate, txType string, lastKnowledge int64) ([]byte, error) {
	q := buildTransactionQuery(sinceDate, txType, lastKnowledge)
	return c.Get(fmt.Sprintf("/budgets/%s/transactions", planID), q)
}

// GetTransactionByID returns a single transaction by ID.
// GET /budgets/{plan_id}/transactions/{transaction_id}
func (c *Client) GetTransactionByID(planID, transactionID string) ([]byte, error) {
	return c.Get(fmt.Sprintf("/budgets/%s/transactions/%s", planID, transactionID), nil)
}

// CreateTransaction creates one or more transactions.
// POST /budgets/{plan_id}/transactions
func (c *Client) CreateTransaction(planID string, body interface{}) ([]byte, error) {
	return c.Post(fmt.Sprintf("/budgets/%s/transactions", planID), body)
}

// UpdateTransactions updates multiple transactions at once.
// PATCH /budgets/{plan_id}/transactions
func (c *Client) UpdateTransactions(planID string, body interface{}) ([]byte, error) {
	return c.Patch(fmt.Sprintf("/budgets/%s/transactions", planID), body)
}

// UpdateTransaction updates a single transaction by ID.
// PUT /budgets/{plan_id}/transactions/{transaction_id}
func (c *Client) UpdateTransaction(planID, transactionID string, body interface{}) ([]byte, error) {
	return c.Put(fmt.Sprintf("/budgets/%s/transactions/%s", planID, transactionID), body)
}

// DeleteTransaction deletes a transaction by ID.
// DELETE /budgets/{plan_id}/transactions/{transaction_id}
func (c *Client) DeleteTransaction(planID, transactionID string) ([]byte, error) {
	return c.Delete(fmt.Sprintf("/budgets/%s/transactions/%s", planID, transactionID))
}

// ImportTransactions triggers a transaction import for linked accounts.
// POST /budgets/{plan_id}/transactions/import
func (c *Client) ImportTransactions(planID string) ([]byte, error) {
	return c.Post(fmt.Sprintf("/budgets/%s/transactions/import", planID), nil)
}

// GetTransactionsByAccount returns transactions for a specific account.
// GET /budgets/{plan_id}/accounts/{account_id}/transactions
func (c *Client) GetTransactionsByAccount(planID, accountID string, sinceDate, txType string, lastKnowledge int64) ([]byte, error) {
	q := buildTransactionQuery(sinceDate, txType, lastKnowledge)
	return c.Get(fmt.Sprintf("/budgets/%s/accounts/%s/transactions", planID, accountID), q)
}

// GetTransactionsByCategory returns transactions for a specific category.
// GET /budgets/{plan_id}/categories/{category_id}/transactions
func (c *Client) GetTransactionsByCategory(planID, categoryID string, sinceDate, txType string, lastKnowledge int64) ([]byte, error) {
	q := buildTransactionQuery(sinceDate, txType, lastKnowledge)
	return c.Get(fmt.Sprintf("/budgets/%s/categories/%s/transactions", planID, categoryID), q)
}

// GetTransactionsByPayee returns transactions for a specific payee.
// GET /budgets/{plan_id}/payees/{payee_id}/transactions
func (c *Client) GetTransactionsByPayee(planID, payeeID string, sinceDate, txType string, lastKnowledge int64) ([]byte, error) {
	q := buildTransactionQuery(sinceDate, txType, lastKnowledge)
	return c.Get(fmt.Sprintf("/budgets/%s/payees/%s/transactions", planID, payeeID), q)
}

// GetTransactionsByMonth returns transactions for a specific month.
// GET /budgets/{plan_id}/months/{month}/transactions
func (c *Client) GetTransactionsByMonth(planID, month string, sinceDate, txType string, lastKnowledge int64) ([]byte, error) {
	q := buildTransactionQuery(sinceDate, txType, lastKnowledge)
	return c.Get(fmt.Sprintf("/budgets/%s/months/%s/transactions", planID, month), q)
}
