package client

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"

	"github.com/thingsdiary/client-go/openapi"
)

func (c *Client) RestoreTopic(ctx context.Context, diaryID string, topicID string) (*Topic, error) {
	if c.credentials == nil {
		return nil, ErrUnauthorized
	}

	key, err := c.getActiveDiaryKey(ctx, diaryID)
	if err != nil {
		return nil, err
	}

	diaryKey, err := decryptWithPrivateKey(
		key.Value,
		c.credentials.EncryptionPrivateKey,
		c.credentials.EncryptionPublicKey,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decrypt diary key")
	}

	url := fmt.Sprintf("%s/v1/diaries/%s/topics/%s/restore", c.baseURL, diaryID, topicID)
	req, err := c.newAuthenticatedRequest(ctx, http.MethodPost, url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}
	req.Header.Set(
		"X-Signature",
		base64.StdEncoding.EncodeToString(signBytes(nil, c.credentials.SigningPrivateKey)),
	)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute request")
	}
	defer resp.Body.Close()

	if err := restoreResponseError(resp, ErrTopicNotFound); err != nil {
		return nil, err
	}

	var apiResponse openapi.GetTopicResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, errors.Wrap(err, "failed to decode response")
	}

	return c.decryptTopic(&apiResponse.Topic, diaryKey)
}
