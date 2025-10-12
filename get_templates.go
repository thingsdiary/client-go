package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"

	"github.com/thingsdiary/client-go/openapi"
)

func (c *Client) GetTemplates(ctx context.Context, diaryID string) ([]*Template, error) {
	if c.credentials == nil {
		return nil, ErrUnauthorized
	}

	// Get the diary to access encryption keys
	diaryData, err := c.getDiary(ctx, diaryID)
	if err != nil {
		return nil, err
	}

	if len(diaryData.EncryptionKeys) == 0 {
		return nil, errors.New("no encryption keys found in diary")
	}

	// Decrypt diary key
	encryptedDiaryKeyValue := diaryData.EncryptionKeys[0].Value
	decryptedDiaryKey, err := decryptWithPrivateKey(
		encryptedDiaryKeyValue,
		c.credentials.EncryptionPrivateKey,
		c.credentials.EncryptionPublicKey,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decrypt diary key")
	}

	// Get templates data
	templatesData, err := c.getTemplates(ctx, diaryID)
	if err != nil {
		return nil, err
	}

	templates := make([]*Template, 0, len(templatesData))
	for _, templateData := range templatesData {
		template, err := c.decryptTemplate(templateData, decryptedDiaryKey)
		if err != nil {
			return nil, err
		}

		templates = append(templates, template)
	}

	return templates, nil
}

func (c *Client) getTemplates(ctx context.Context, diaryID string) ([]*openapi.Template, error) {
	var url = fmt.Sprintf("%s/api/v1/diaries/%s/templates", c.baseURL, diaryID)
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

	var apiResponse openapi.GetTemplatesResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, errors.Wrap(err, "failed to decode response")
	}

	return apiResponse.Templates, nil
}
