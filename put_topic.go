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

// PutTopicParams contains parameters for creating/updating a topic
type PutTopicParams struct {
	Title             string
	Description       string
	Color             string
	DefaultTemplateID mo.Option[string]
}

// GetTopicDetails extracts topic details from parameters
func (p PutTopicParams) GetTopicDetails() TopicDetails {
	return TopicDetails{
		Title:       p.Title,
		Description: p.Description,
		Color:       p.Color,
	}
}

// PutTopic creates or updates a topic in a diary
func (c *Client) PutTopic(ctx context.Context, diaryID, topicID string, params PutTopicParams) (*Topic, error) {
	if c.credentials == nil {
		return nil, ErrUnauthorized
	}

	// Get diary to access encryption keys
	diaryData, err := c.getDiary(ctx, diaryID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get diary")
	}

	if len(diaryData.EncryptionKeys) == 0 {
		return nil, errors.New("no encryption keys found in diary")
	}

	diaryKeyID := diaryData.EncryptionKeys[0].Id

	// Generate entity key for topic encryption
	entityKey, err := generateSymmetricKey()
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate entity key")
	}

	// Encrypt topic details
	topicDetails := params.GetTopicDetails()
	topicDetailsJSON, err := json.Marshal(topicDetails)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal topic details")
	}

	detailsNonce, encryptedDetails, err := encryptWithSymmetricKey(topicDetailsJSON, entityKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to encrypt topic details")
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

	// Encrypt entity key with diary key
	keyNonce, encryptedEntityKey, err := encryptWithSymmetricKey(entityKey, decryptedDiaryKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to encrypt entity key")
	}

	var apiDefaultTemplateID mo.Option[openapi.TemplateID]
	if params.DefaultTemplateID.IsPresent() {
		apiDefaultTemplateID = mo.Some(openapi.TemplateID(params.DefaultTemplateID.MustGet()))
	}

	// Create request
	request := openapi.PutTopicRequest{
		Version:           NewVersion(),
		DefaultTemplateId: apiDefaultTemplateID,
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

	url := fmt.Sprintf("%s/api/v1/diaries/%s/topics/%s", c.baseURL, diaryID, topicID)
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

	var apiResponse openapi.PutTopicResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, errors.Wrap(err, "failed to decode response")
	}

	// Decrypt and return topic
	return c.decryptTopic(&apiResponse.Topic, decryptedDiaryKey)
}

// decryptTopic decrypts an encrypted topic to plaintext
func (c *Client) decryptTopic(apiTopic *openapi.Topic, diaryKey []byte) (*Topic, error) {
	// Decrypt entity key
	decryptedEntityKey, err := decryptWithSymmetricKey(
		apiTopic.Encryption.EncryptedKeyNonce,
		apiTopic.Encryption.EncryptedKeyData,
		diaryKey,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decrypt entity key")
	}

	// Decrypt topic details
	decryptedDetailsBytes, err := decryptWithSymmetricKey(
		apiTopic.Details.Nonce,
		apiTopic.Details.Data,
		decryptedEntityKey,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decrypt topic details")
	}

	var topicDetails TopicDetails
	if err := json.Unmarshal(decryptedDetailsBytes, &topicDetails); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal topic details")
	}

	var defaultTemplateID mo.Option[string]
	if apiTopic.DefaultTemplateId.IsPresent() {
		defaultTemplateID = mo.Some(string(apiTopic.DefaultTemplateId.MustGet()))
	}

	topic := &Topic{
		ID:                apiTopic.Id,
		DiaryID:           string(apiTopic.DiaryId),
		Title:             topicDetails.Title,
		Description:       topicDetails.Description,
		Color:             topicDetails.Color,
		DefaultTemplateID: defaultTemplateID,
		CreatedAt:         apiTopic.CreatedAt,
		UpdatedAt:         apiTopic.UpdatedAt,
		DeletedAt:         apiTopic.DeletedAt,
		Version:           apiTopic.Version,
	}

	return topic, nil
}
