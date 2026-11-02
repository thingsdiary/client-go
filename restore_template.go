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

func (c *Client) RestoreTemplate(ctx context.Context, diaryID string, templateID string) (*Template, error) {
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

	url := fmt.Sprintf("%s/v1/diaries/%s/templates/%s/restore", c.baseURL, diaryID, templateID)
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

	if err := restoreResponseError(resp, ErrTemplateNotFound); err != nil {
		return nil, err
	}

	var apiResponse openapi.GetTemplateResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, errors.Wrap(err, "failed to decode response")
	}

	return c.decryptTemplate(&apiResponse.Template, diaryKey)
}
