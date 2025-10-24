package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"

	"github.com/thingsdiary/client-go/openapi"
)

func (c *Client) GetTemplateByID(ctx context.Context, diaryID, templateID string) (*Template, error) {
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

	// Get template from API
	apiTemplate, err := c.getTemplate(ctx, diaryID, templateID)
	if err != nil {
		return nil, err
	}

	// Decrypt template
	return c.decryptTemplate(apiTemplate, decryptedDiaryKey)
}

func (c *Client) getTemplate(ctx context.Context, diaryID, templateID string) (*openapi.Template, error) {
	url := fmt.Sprintf("%s/api/v1/diaries/%s/templates/%s", c.baseURL, diaryID, templateID)
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
		return nil, ErrTemplateNotFound
	}

	if resp.StatusCode == http.StatusForbidden {
		return nil, ErrForbidden
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var apiResponse openapi.GetTemplateResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, errors.Wrap(err, "failed to decode response")
	}

	return &apiResponse.Template, nil
}

// decryptTemplate decrypts an encrypted template to plaintext
func (c *Client) decryptTemplate(apiTemplate *openapi.Template, diaryKey []byte) (*Template, error) {
	// Decrypt entity key
	decryptedEntityKey, err := decryptWithSymmetricKey(
		apiTemplate.Encryption.EncryptedKeyNonce,
		apiTemplate.Encryption.EncryptedKeyData,
		diaryKey,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decrypt entity key")
	}

	// Decrypt template details
	decryptedDetailsBytes, err := decryptWithSymmetricKey(
		apiTemplate.Details.Nonce,
		apiTemplate.Details.Data,
		decryptedEntityKey,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decrypt template details")
	}

	var templateDetails TemplateDetails
	if err := json.Unmarshal(decryptedDetailsBytes, &templateDetails); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal template details")
	}

	template := Template{
		ID:        apiTemplate.Id,
		DiaryID:   string(apiTemplate.DiaryId),
		Content:   templateDetails.Content,
		CreatedAt: apiTemplate.CreatedAt,
		UpdatedAt: apiTemplate.UpdatedAt,
		DeletedAt: apiTemplate.DeletedAt,
		Version:   apiTemplate.Version,
	}

	return &template, nil
}
