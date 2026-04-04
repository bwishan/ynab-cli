package api

import (
	"fmt"
	"net/url"
	"strconv"
)

// GetPlans returns all plans (budgets) for the authenticated user.
// GET /plans?include_accounts=true
func (c *Client) GetPlans(includeAccounts bool) ([]byte, error) {
	q := url.Values{}
	if includeAccounts {
		q.Set("include_accounts", "true")
	}
	return c.Get("/budgets", q)
}

// GetPlanByID returns a single plan by ID.
// GET /plans/{plan_id}?last_knowledge_of_server=N
func (c *Client) GetPlanByID(planID string, lastKnowledge int64) ([]byte, error) {
	q := url.Values{}
	if lastKnowledge > 0 {
		q.Set("last_knowledge_of_server", strconv.FormatInt(lastKnowledge, 10))
	}
	return c.Get(fmt.Sprintf("/budgets/%s", planID), q)
}

// GetPlanSettings returns the settings for a plan.
// GET /plans/{plan_id}/settings
func (c *Client) GetPlanSettings(planID string) ([]byte, error) {
	return c.Get(fmt.Sprintf("/budgets/%s/settings", planID), nil)
}
