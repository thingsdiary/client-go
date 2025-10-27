package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"

	"github.com/thingsdiary/client-go/openapi"
)

func (c *Client) GetDiaryByID(ctx context.Context, id string) (*Diary, error) {
	if c.credentials == nil {
		return nil, ErrUnauthorized
	}

	diaryData, err := c.getDiary(ctx, id)
	if err != nil {
		return nil, err
	}

	return c.decryptDiary(diaryData)
}

func (c *Client) getDiary(ctx context.Context, id string) (*openapi.Diary, error) {
	var url = fmt.Sprintf("%s/v1/diaries/%s", c.baseURL, id)
	req, err := c.newAuthenticatedRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute request")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrDiaryNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var apiResponse openapi.GetDiaryResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, errors.Wrap(err, "failed to decode response")
	}

	return &apiResponse.Diary, nil
}
