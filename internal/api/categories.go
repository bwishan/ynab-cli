package api

import (
	"fmt"
	"net/url"
)

// GetCategories returns all categories for a plan.
// GET /budgets/{plan_id}/categories
func (c *Client) GetCategories(planID string, lastKnowledge int64) ([]byte, error) {
	var query url.Values
	if lastKnowledge > 0 {
		query = url.Values{}
		query.Set("last_knowledge_of_server", fmt.Sprintf("%d", lastKnowledge))
	}
	return c.Get(fmt.Sprintf("/budgets/%s/categories", planID), query)
}

// GetCategoryByID returns a single category.
// GET /budgets/{plan_id}/categories/{category_id}
func (c *Client) GetCategoryByID(planID, categoryID string) ([]byte, error) {
	return c.Get(fmt.Sprintf("/budgets/%s/categories/%s", planID, categoryID), nil)
}

// CreateCategory creates a new category in the plan.
// POST /budgets/{plan_id}/categories
func (c *Client) CreateCategory(planID string, body interface{}) ([]byte, error) {
	return c.Post(fmt.Sprintf("/budgets/%s/categories", planID), body)
}

// UpdateCategory updates an existing category.
// PATCH /budgets/{plan_id}/categories/{category_id}
func (c *Client) UpdateCategory(planID, categoryID string, body interface{}) ([]byte, error) {
	return c.Patch(fmt.Sprintf("/budgets/%s/categories/%s", planID, categoryID), body)
}

// GetMonthCategory returns a single category for a specific month.
// GET /budgets/{plan_id}/months/{month}/categories/{category_id}
func (c *Client) GetMonthCategory(planID, month, categoryID string) ([]byte, error) {
	return c.Get(fmt.Sprintf("/budgets/%s/months/%s/categories/%s", planID, month, categoryID), nil)
}

// UpdateMonthCategory updates a category for a specific month (e.g., budgeted amount).
// PATCH /budgets/{plan_id}/months/{month}/categories/{category_id}
func (c *Client) UpdateMonthCategory(planID, month, categoryID string, body interface{}) ([]byte, error) {
	return c.Patch(fmt.Sprintf("/budgets/%s/months/%s/categories/%s", planID, month, categoryID), body)
}

// CreateCategoryGroup creates a new category group in the plan.
// POST /budgets/{plan_id}/category_groups
func (c *Client) CreateCategoryGroup(planID string, body interface{}) ([]byte, error) {
	return c.Post(fmt.Sprintf("/budgets/%s/category_groups", planID), body)
}

// UpdateCategoryGroup updates an existing category group.
// PATCH /budgets/{plan_id}/category_groups/{category_group_id}
func (c *Client) UpdateCategoryGroup(planID, groupID string, body interface{}) ([]byte, error) {
	return c.Patch(fmt.Sprintf("/budgets/%s/category_groups/%s", planID, groupID), body)
}
