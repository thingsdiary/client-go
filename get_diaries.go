package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"

	"github.com/thingsdiary/client-go/openapi"
)

func (c *Client) GetDiaries(ctx context.Context) ([]*Diary, error) {
	if c.credentials == nil {
		return nil, ErrUnauthorized
	}

	diariesData, err := c.getDiaries(ctx)
	if err != nil {
		return nil, err
	}

	diaries := make([]*Diary, 0, len(diariesData))
	for _, diaryData := range diariesData {
		diary, err := c.decryptDiary(diaryData)
		if err != nil {
			return nil, err
		}
		diaries = append(diaries, diary)
	}

	return diaries, nil
}

func (c *Client) getDiaries(ctx context.Context) ([]*openapi.Diary, error) {
	var url = fmt.Sprintf("%s/v1/diaries", c.baseURL)
	req, err := c.newAuthenticatedRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute request")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrUnauthorized
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var apiResponse openapi.GetDiariesResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, errors.Wrap(err, "failed to decode response")
	}

	return apiResponse.Diaries, nil
}
