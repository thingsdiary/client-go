package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"

	"github.com/thingsdiary/client-go/openapi"
)

func (c *Client) getActiveDiaryKey(ctx context.Context, diaryID string) (*openapi.DiaryEncryptionKey, error) {
	// TODO: Implement client-side caching for keys (TTL ~5min)
	url := fmt.Sprintf("%s/api/v1/diaries/%s/keys", c.baseURL, diaryID)

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

	if resp.StatusCode == http.StatusForbidden {
		return nil, ErrForbidden
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("unexpected status code: %s", resp.Status)
	}

	var apiResponse openapi.GetDiaryKeysResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, errors.Wrap(err, "failed to decode response")
	}

	for i := range apiResponse.Keys {
		if apiResponse.Keys[i].Status == openapi.Active {
			return &apiResponse.Keys[i], nil
		}
	}

	return nil, errors.New("no active encryption key found")
}
