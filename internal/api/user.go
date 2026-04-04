package api

// GetUser returns the authenticated user's information.
// GET /user
func (c *Client) GetUser() ([]byte, error) {
	return c.Get("/user", nil)
}
