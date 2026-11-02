package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/pkg/errors"
)

type DeleteTemplateParams struct {
	Soft *bool
}

func (c *Client) DeleteTemplate(
	ctx context.Context,
	diaryID string,
	templateID string,
	params ...DeleteTemplateParams,
) error {
	if c.credentials == nil {
		return ErrUnauthorized
	}

	var p DeleteTemplateParams
	if len(params) > 0 {
		p = params[0]
	}

	urlStr := fmt.Sprintf("%s/v1/diaries/%s/templates/%s", c.baseURL, diaryID, templateID)
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
		return ErrTemplateNotFound
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
