package client

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/samber/mo"

	v1 "github.com/thingsdiary/client-go/proto"
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

	// Get encryption keys
	key, err := c.getActiveDiaryKey(ctx, diaryID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get active diary key")
	}

	diaryKeyID := key.Id

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

	encryption := &v1.Encryption{
		DiaryKeyId:        diaryKeyID,
		EncryptedKeyNonce: keyNonce,
		EncryptedKeyData:  encryptedEntityKey,
	}

	details := &v1.EncryptedData{Nonce: detailsNonce, Data: encryptedDetails}

	if err := validateProtoEncryption(encryption); err != nil {
		return nil, errors.Wrap(err, "request validation failed")
	}

	if err := validateProtoEncryptedData(details, "details"); err != nil {
		return nil, errors.Wrap(err, "request validation failed")
	}

	req := &v1.PutTopicRequest{
		DiaryId:           diaryID,
		TopicId:           topicID,
		Version:           NewVersion(),
		Encryption:        encryption,
		Details:           details,
		DefaultTemplateId: protoString(params.DefaultTemplateID),
	}

	response := &v1.Topic{}
	if err := c.rpc(ctx, "io.thingsdiary.v1.topic.put", req, response); err != nil {
		return nil, err
	}

	return topicFromProto(response, decryptedDiaryKey)
}
