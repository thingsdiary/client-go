package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

type DeleteTopicParams struct {
	DeleteEntries bool
}

func (c *Client) DeleteTopic(ctx context.Context, diaryID string, topicID string, params ...DeleteTopicParams) error {
	if c.credentials == nil {
		return ErrUnauthorized
	}

	var p DeleteTopicParams
	if len(params) > 0 {
		p = params[0]
	}

	urlStr := fmt.Sprintf("%s/api/v1/diaries/%s/topics/%s", c.baseURL, diaryID, topicID)
	u, err := url.Parse(urlStr)
	if err != nil {
		return errors.Wrap(err, "failed to parse URL")
	}

	if p.DeleteEntries {
		q := url.Values{}
		q.Set("delete_entries", "true")
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
		return ErrTopicNotFound
	}

	if resp.StatusCode == http.StatusForbidden {
		return ErrForbidden
	}

	if resp.StatusCode != http.StatusNoContent {
		return errors.Errorf("unexpected status code: %s", resp.Status)
	}

	return nil
}
