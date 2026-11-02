package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/pkg/errors"
)

type DeleteEntryParams struct {
	Soft *bool
}

func (c *Client) DeleteEntry(ctx context.Context, diaryID string, entryID string, params ...DeleteEntryParams) error {
	if c.credentials == nil {
		return ErrUnauthorized
	}

	var p DeleteEntryParams
	if len(params) > 0 {
		p = params[0]
	}

	urlStr := fmt.Sprintf("%s/v1/diaries/%s/entries/%s", c.baseURL, diaryID, entryID)
	u, err := url.Parse(urlStr)
	if err != nil {
		return errors.Wrap(err, "failed to parse URL")
	}

	if p.Soft != nil {
		q := u.Query()
		q.Set("soft", strconv.FormatBool(*p.Soft))
		u.RawQuery = q.Encode()
	}

	req, err := c.newAuthenticatedRequest(ctx, http.MethodDelete, u.String(), nil)
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

	if resp.StatusCode == http.StatusGone {
		return ErrEntityRemoved
	}

	if resp.StatusCode != http.StatusNoContent {
		return errors.Errorf("unexpected status code: %s", resp.Status)
	}

	return nil
}
