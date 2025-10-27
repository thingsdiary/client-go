package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"

	"github.com/thingsdiary/client-go/openapi"
)

func (c *Client) GetTopics(ctx context.Context, diaryID string) ([]*Topic, error) {
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

	// Get topics data
	topicsData, err := c.getTopics(ctx, diaryID)
	if err != nil {
		return nil, err
	}

	topics := make([]*Topic, 0, len(topicsData))
	for _, topicData := range topicsData {
		topic, err := c.decryptTopic(topicData, decryptedDiaryKey)
		if err != nil {
			return nil, err
		}

		topics = append(topics, topic)
	}

	return topics, nil
}

func (c *Client) getTopics(ctx context.Context, diaryID string) ([]*openapi.Topic, error) {
	var url = fmt.Sprintf("%s/v1/diaries/%s/topics", c.baseURL, diaryID)
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

	var apiResponse openapi.GetTopicsResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, errors.Wrap(err, "failed to decode response")
	}

	return apiResponse.Topics, nil
}
