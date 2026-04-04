package api

import "fmt"

// GetMoneyMovements returns all money movements for a plan.
// GET /plans/{plan_id}/money_movements
func (c *Client) GetMoneyMovements(planID string) ([]byte, error) {
	return c.Get(fmt.Sprintf("/budgets/%s/money_movements", planID), nil)
}

// GetMoneyMovementsByMonth returns money movements for a specific month.
// GET /plans/{plan_id}/months/{month}/money_movements
func (c *Client) GetMoneyMovementsByMonth(planID, month string) ([]byte, error) {
	return c.Get(fmt.Sprintf("/budgets/%s/months/%s/money_movements", planID, month), nil)
}

// GetMoneyMovementGroups returns all money movement groups for a plan.
// GET /plans/{plan_id}/money_movement_groups
func (c *Client) GetMoneyMovementGroups(planID string) ([]byte, error) {
	return c.Get(fmt.Sprintf("/budgets/%s/money_movement_groups", planID), nil)
}

// GetMoneyMovementGroupsByMonth returns money movement groups for a specific month.
// GET /plans/{plan_id}/months/{month}/money_movement_groups
func (c *Client) GetMoneyMovementGroupsByMonth(planID, month string) ([]byte, error) {
	return c.Get(fmt.Sprintf("/budgets/%s/months/%s/money_movement_groups", planID, month), nil)
}
