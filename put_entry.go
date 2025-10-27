package client

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/samber/mo"

	"github.com/thingsdiary/client-go/openapi"
)

// PutEntryParams contains parameters for creating/updating an entry
type PutEntryParams struct {
	Content       string
	TopicID       mo.Option[string]
	Archived      bool
	Bookmarked    bool
	PreviewHidden bool
}

// GetEntryDetails extracts entry details from parameters
func (p PutEntryParams) GetEntryDetails() EntryDetails {
	return EntryDetails{
		Content:       p.Content,
		Archived:      p.Archived,
		Bookmarked:    p.Bookmarked,
		PreviewHidden: p.PreviewHidden,
	}
}

// GetEntryPreview extracts preview content (same as details for now)
func (p PutEntryParams) GetEntryPreview() EntryDetails {
	return p.GetEntryDetails()
}

// PutEntry creates or updates an entry in a diary
func (c *Client) PutEntry(ctx context.Context, diaryID, entryID string, params PutEntryParams) (*Entry, error) {
	if c.credentials == nil {
		return nil, ErrUnauthorized
	}

	// Get encryption keys
	key, err := c.getActiveDiaryKey(ctx, diaryID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get active diary key")
	}

	diaryKeyID := key.Id

	// Generate entity key for entry encryption
	entityKey, err := generateSymmetricKey()
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate entity key")
	}

	// Encrypt entry details
	entryDetails := params.GetEntryDetails()
	entryDetailsJSON, err := json.Marshal(entryDetails)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal entry details")
	}

	detailsNonce, encryptedDetails, err := encryptWithSymmetricKey(entryDetailsJSON, entityKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to encrypt entry details")
	}

	// Encrypt entry preview (same as details for now)
	entryPreview := params.GetEntryPreview()
	entryPreviewJSON, err := json.Marshal(entryPreview)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal entry preview")
	}

	previewNonce, encryptedPreview, err := encryptWithSymmetricKey(entryPreviewJSON, entityKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to encrypt entry preview")
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

	// Convert TopicID if present
	var topicID mo.Option[openapi.TopicID]
	if params.TopicID.IsPresent() {
		topicID = mo.Some(openapi.TopicID(params.TopicID.MustGet()))
	}

	// Create request
	request := openapi.PutEntryRequest{
		Version: NewVersion(),
		TopicId: topicID,
		Encryption: openapi.DiaryEncryption{
			DiaryKeyId:        diaryKeyID,
			EncryptedKeyNonce: keyNonce,
			EncryptedKeyData:  encryptedEntityKey,
		},
		Details: openapi.EncryptedData{
			Nonce: detailsNonce,
			Data:  encryptedDetails,
		},
		Preview: openapi.EncryptedData{
			Nonce: previewNonce,
			Data:  encryptedPreview,
		},
	}

	// Validate request before sending
	if err := request.Validate(); err != nil {
		return nil, errors.Wrap(err, "request validation failed")
	}

	url := fmt.Sprintf("%s/v1/diaries/%s/entries/%s", c.baseURL, diaryID, entryID)
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

	if resp.StatusCode == http.StatusBadRequest {
		var errorResp openapi.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil {
			if errorResp.ErrorCode == openapi.ResponseErrorCodeTopicNotFound {
				return nil, errors.Wrap(ErrTopicNotFound, "could not put entry")
			}
		}

		return nil, errors.Errorf("bad request: %s", resp.Status)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("unexpected status code: %s", resp.Status)
	}

	var apiResponse openapi.PutEntryResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, errors.Wrap(err, "failed to decode response")
	}

	// Decrypt and return entry
	return c.decryptEntry(&apiResponse.Entry, decryptedDiaryKey)
}

// decryptEntry decrypts an encrypted entry to plaintext
func (c *Client) decryptEntry(apiEntry *openapi.Entry, diaryKey []byte) (*Entry, error) {
	// Decrypt entity key
	decryptedEntityKey, err := decryptWithSymmetricKey(
		apiEntry.Encryption.EncryptedKeyNonce,
		apiEntry.Encryption.EncryptedKeyData,
		diaryKey,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decrypt entity key")
	}

	// Decrypt entry details
	decryptedDetailsBytes, err := decryptWithSymmetricKey(
		apiEntry.Details.Nonce,
		apiEntry.Details.Data,
		decryptedEntityKey,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decrypt entry details")
	}

	var entryDetails EntryDetails
	if err := json.Unmarshal(decryptedDetailsBytes, &entryDetails); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal entry details")
	}

	// Convert TopicID
	var topicID mo.Option[string]
	if apiEntry.TopicId.IsPresent() {
		topicID = mo.Some(string(apiEntry.TopicId.MustGet()))
	}

	entry := Entry{
		ID:            apiEntry.Id,
		DiaryID:       string(apiEntry.DiaryId),
		Content:       entryDetails.Content,
		TopicID:       topicID,
		Archived:      entryDetails.Archived,
		Bookmarked:    entryDetails.Bookmarked,
		PreviewHidden: entryDetails.PreviewHidden,
		CreatedAt:     apiEntry.CreatedAt,
		UpdatedAt:     apiEntry.UpdatedAt,
		DeletedAt:     apiEntry.DeletedAt,
		Version:       apiEntry.Version,
	}

	return &entry, nil
}
