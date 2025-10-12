package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

func (c *Client) Logout(ctx context.Context) error {
	url := fmt.Sprintf("%s/api/v1/auth/logout", c.baseURL)
	req, err := c.newAuthenticatedRequest(ctx, http.MethodPost, url, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return errors.New("unauthorized")
	}

	if resp.StatusCode != http.StatusNoContent {
		return errors.Errorf("logout failed: %s", resp.Status)
	}

	c.authToken = ""
	c.credentials = nil

	return nil
}
