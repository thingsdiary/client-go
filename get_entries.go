package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"

	"github.com/thingsdiary/client-go/openapi"
)

func (c *Client) GetEntries(ctx context.Context, diaryID string) ([]*Entry, error) {
	if c.credentials == nil {
		return nil, ErrUnauthorized
	}

	// Get encryption keys
	key, err := c.getActiveDiaryKey(ctx, diaryID)
	if err != nil {
		return nil, err
	}

	// Decrypt diary key
	decryptedDiaryKey, err := decryptWithPrivateKey(
		key.Value,
		c.credentials.EncryptionPrivateKey,
		c.credentials.EncryptionPublicKey,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decrypt diary key")
	}

	// Get entries data
	entriesData, err := c.getEntries(ctx, diaryID)
	if err != nil {
		return nil, err
	}

	entries := make([]*Entry, 0, len(entriesData))
	for _, entryData := range entriesData {
		entry, err := c.decryptEntry(entryData, decryptedDiaryKey)
		if err != nil {
			return nil, err
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

func (c *Client) getEntries(ctx context.Context, diaryID string) ([]*openapi.Entry, error) {
	var url = fmt.Sprintf("%s/v1/diaries/%s/entries", c.baseURL, diaryID)
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

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrDiaryNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var apiResponse openapi.GetEntriesResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, errors.Wrap(err, "failed to decode response")
	}

	return apiResponse.Entries, nil
}
