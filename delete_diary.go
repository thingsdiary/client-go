package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

func (c *Client) DeleteDiary(ctx context.Context, diaryID string) error {
	if c.credentials == nil {
		return ErrUnauthorized
	}

	url := fmt.Sprintf("%s/v1/diaries/%s", c.baseURL, diaryID)
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
		return ErrDiaryNotFound
	}

	if resp.StatusCode == http.StatusForbidden {
		return ErrForbidden
	}

	if resp.StatusCode != http.StatusNoContent {
		return errors.Errorf("unexpected status code: %s", resp.Status)
	}

	return nil
}
