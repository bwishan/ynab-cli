package api

import (
	"fmt"
	"net/url"
	"strconv"
)

// GetMonths returns all budget months for a plan.
// GET /plans/{plan_id}/months?last_knowledge_of_server=N
func (c *Client) GetMonths(planID string, lastKnowledge int64) ([]byte, error) {
	q := url.Values{}
	if lastKnowledge > 0 {
		q.Set("last_knowledge_of_server", strconv.FormatInt(lastKnowledge, 10))
	}
	return c.Get(fmt.Sprintf("/budgets/%s/months", planID), q)
}

// GetMonth returns a single budget month.
// GET /plans/{plan_id}/months/{month}
func (c *Client) GetMonth(planID, month string) ([]byte, error) {
	return c.Get(fmt.Sprintf("/budgets/%s/months/%s", planID, month), nil)
}
