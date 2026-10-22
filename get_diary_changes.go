package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/samber/mo"
	"github.com/thingsdiary/client-go/openapi"
)

type GetDiaryChangesOptions struct {
	NextPageToken mo.Option[string]
	// Filter changes created after this timestamp
	Since mo.Option[time.Time]
	// Range: 1-500, default: 100
	Limit mo.Option[int]
}

// GetDiaryChanges returns diary changes for incremental synchronization.
func (c *Client) GetDiaryChanges(ctx context.Context, diaryID string, opts GetDiaryChangesOptions) (*GetDiaryChangesResult, error) {
	if c.credentials == nil {
		return nil, ErrUnauthorized
	}

	apiChanges, nextPageToken, err := c.getDiaryChanges(ctx, diaryID, opts)
	if err != nil {
		return nil, err
	}

	changes := make([]*DiaryChange, len(apiChanges))
	for i, apiChange := range apiChanges {
		changes[i] = &DiaryChange{
			ID:         apiChange.Id,
			CreatedAt:  apiChange.CreatedAt,
			Operation:  ChangeOperation(apiChange.Operation),
			EntityType: ChangeEntityType(apiChange.EntityType),
			EntityID:   apiChange.EntityId,
		}
	}

	return &GetDiaryChangesResult{
		Changes:       changes,
		NextPageToken: nextPageToken,
	}, nil
}

func (c *Client) getDiaryChanges(ctx context.Context, diaryID string, opts GetDiaryChangesOptions) ([]*openapi.DiaryChange, string, error) {
	baseURL := fmt.Sprintf("%s/v1/diaries/%s/changes", c.baseURL, diaryID)

	params := url.Values{}
	if token, ok := opts.NextPageToken.Get(); ok {
		params.Set("next_page_token", token)
	}
	if since, ok := opts.Since.Get(); ok {
		params.Set("since", since.Format(time.RFC3339Nano))
	}
	if limit, ok := opts.Limit.Get(); ok {
		params.Set("limit", strconv.Itoa(limit))
	}

	requestURL := baseURL
	if len(params) > 0 {
		requestURL = baseURL + "?" + params.Encode()
	}

	req, err := c.newAuthenticatedRequest(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, "", errors.Wrap(err, "failed to create request")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, "", errors.Wrap(err, "failed to execute request")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, "", ErrUnauthorized
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, "", ErrDiaryNotFound
	}

	if resp.StatusCode == http.StatusBadRequest {
		return nil, "", ErrBadRequest
	}

	if resp.StatusCode != http.StatusOK {
		return nil, "", errors.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var apiResponse openapi.GetDiaryChangesResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, "", errors.Wrap(err, "failed to decode response")
	}

	nextPageToken := ""
	if token, ok := apiResponse.NextPageToken.Get(); ok {
		nextPageToken = token
	}

	return apiResponse.Changes, nextPageToken, nil
}
