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

// PutTemplateParams contains parameters for creating/updating a template
type PutTemplateParams struct {
	Content string
}

// GetTemplateDetails extracts template details from parameters
func (p PutTemplateParams) GetTemplateDetails() TemplateDetails {
	return TemplateDetails(p)
}

// PutTemplate creates or updates a template in a diary
func (c *Client) PutTemplate(ctx context.Context, diaryID, templateID string, params PutTemplateParams) (*Template, error) {
	if c.credentials == nil {
		return nil, ErrUnauthorized
	}

	// Get encryption keys
	key, err := c.getActiveDiaryKey(ctx, diaryID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get active diary key")
	}

	diaryKeyID := key.Id

	// Generate entity key for template encryption
	entityKey, err := generateSymmetricKey()
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate entity key")
	}

	// Encrypt template details
	templateDetails := params.GetTemplateDetails()
	templateDetailsJSON, err := json.Marshal(templateDetails)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal template details")
	}

	detailsNonce, encryptedDetails, err := encryptWithSymmetricKey(templateDetailsJSON, entityKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to encrypt template details")
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

	// Encrypt entity key with diary key
	keyNonce, encryptedEntityKey, err := encryptWithSymmetricKey(entityKey, decryptedDiaryKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to encrypt entity key")
	}

	// Create request
	request := openapi.PutTemplateRequest{
		Version: NewVersion(),
		Encryption: openapi.DiaryEncryption{
			DiaryKeyId:        diaryKeyID,
			EncryptedKeyNonce: keyNonce,
			EncryptedKeyData:  encryptedEntityKey,
		},
		Details: openapi.EncryptedData{
			Nonce: detailsNonce,
			Data:  encryptedDetails,
		},
	}

	// Validate request before sending
	if err := request.Validate(); err != nil {
		return nil, errors.Wrap(err, "request validation failed")
	}

	url := fmt.Sprintf("%s/api/v1/diaries/%s/templates/%s", c.baseURL, diaryID, templateID)
	req, err := c.newAuthenticatedRequest(ctx, http.MethodPut, url, request)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}

	requestJSON, err := json.Marshal(request)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal request")
	}

	signature := signBytes(requestJSON, c.credentials.SigningPrivateKey)
	req.Header.Set("X-Signature", base64.StdEncoding.EncodeToString(signature))

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

	var apiResponse openapi.PutTemplateResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, errors.Wrap(err, "failed to decode response")
	}

	// Decrypt and return template
	return c.decryptTemplate(&apiResponse.Template, decryptedDiaryKey)
}
