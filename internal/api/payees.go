package api

import (
	"fmt"
	"net/url"
	"strconv"
)

// GetPayees returns all payees for a plan.
// GET /plans/{plan_id}/payees?last_knowledge_of_server=N
func (c *Client) GetPayees(planID string, lastKnowledge int64) ([]byte, error) {
	q := url.Values{}
	if lastKnowledge > 0 {
		q.Set("last_knowledge_of_server", strconv.FormatInt(lastKnowledge, 10))
	}
	return c.Get(fmt.Sprintf("/budgets/%s/payees", planID), q)
}

// GetPayeeByID returns a single payee by ID.
// GET /plans/{plan_id}/payees/{payee_id}
func (c *Client) GetPayeeByID(planID, payeeID string) ([]byte, error) {
	return c.Get(fmt.Sprintf("/budgets/%s/payees/%s", planID, payeeID), nil)
}

// CreatePayee creates a new payee in a plan.
// POST /plans/{plan_id}/payees  body: { payee: { name } }
func (c *Client) CreatePayee(planID string, body interface{}) ([]byte, error) {
	return c.Post(fmt.Sprintf("/budgets/%s/payees", planID), body)
}

// UpdatePayee updates an existing payee.
// PATCH /plans/{plan_id}/payees/{payee_id}  body: { payee: { name } }
func (c *Client) UpdatePayee(planID, payeeID string, body interface{}) ([]byte, error) {
	return c.Patch(fmt.Sprintf("/budgets/%s/payees/%s", planID, payeeID), body)
}

// GetPayeeLocations returns all payee locations for a plan.
// GET /plans/{plan_id}/payee_locations
func (c *Client) GetPayeeLocations(planID string) ([]byte, error) {
	return c.Get(fmt.Sprintf("/budgets/%s/payee_locations", planID), nil)
}

// GetPayeeLocationByID returns a single payee location by ID.
// GET /plans/{plan_id}/payee_locations/{payee_location_id}
func (c *Client) GetPayeeLocationByID(planID, payeeLocationID string) ([]byte, error) {
	return c.Get(fmt.Sprintf("/budgets/%s/payee_locations/%s", planID, payeeLocationID), nil)
}

// GetPayeeLocationsByPayee returns all locations for a specific payee.
// GET /plans/{plan_id}/payees/{payee_id}/payee_locations
func (c *Client) GetPayeeLocationsByPayee(planID, payeeID string) ([]byte, error) {
	return c.Get(fmt.Sprintf("/budgets/%s/payees/%s/payee_locations", planID, payeeID), nil)
}
