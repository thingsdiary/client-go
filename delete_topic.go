package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/pkg/errors"
)

type DeleteTopicParams struct {
	DeleteEntries bool
	Soft          *bool
}

func (c *Client) DeleteTopic(ctx context.Context, diaryID string, topicID string, params ...DeleteTopicParams) error {
	if c.credentials == nil {
		return ErrUnauthorized
	}

	var p DeleteTopicParams
	if len(params) > 0 {
		p = params[0]
	}

	urlStr := fmt.Sprintf("%s/v1/diaries/%s/topics/%s", c.baseURL, diaryID, topicID)
	u, err := url.Parse(urlStr)
	if err != nil {
		return errors.Wrap(err, "failed to parse URL")
	}

	q := url.Values{}
	if p.DeleteEntries {
		q.Set("delete_entries", "true")
	}
	if p.Soft != nil {
		q.Set("soft", strconv.FormatBool(*p.Soft))
	}
	u.RawQuery = q.Encode()

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

	if resp.StatusCode == http.StatusGone {
		return ErrEntityRemoved
	}

	if resp.StatusCode != http.StatusNoContent {
		return errors.Errorf("unexpected status code: %s", resp.Status)
	}

	return nil
}
