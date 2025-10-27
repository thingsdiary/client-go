package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

func (c *Client) DeleteEntry(ctx context.Context, diaryID string, entryID string) error {
	if c.credentials == nil {
		return ErrUnauthorized
	}

	url := fmt.Sprintf("%s/v1/diaries/%s/entries/%s", c.baseURL, diaryID, entryID)
	req, err := c.newAuthenticatedRequest(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return errors.Wrap(err, "failed to create request")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to execute request")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return ErrEntryNotFound
	}

	if resp.StatusCode == http.StatusForbidden {
		return ErrForbidden
	}

	if resp.StatusCode != http.StatusNoContent {
		return errors.Errorf("unexpected status code: %s", resp.Status)
	}

	return nil
}
