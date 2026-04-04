package api

// SetBaseURL overrides the base URL for testing purposes.
// This allows external test packages to point the client at a test server.
func (c *Client) SetBaseURL(url string) {
	c.baseURL = url
}
