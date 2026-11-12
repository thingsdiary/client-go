package client

import (
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/samber/mo"
	v1 "github.com/thingsdiary/client-go/proto"
	"google.golang.org/protobuf/proto"
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

	diaryKey, err := decryptWithPrivateKey(
		key.Value,
		c.credentials.EncryptionPrivateKey,
		c.credentials.EncryptionPublicKey,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decrypt diary key")
	}

	current, err := c.getTopic(ctx, diaryID, topicID)
	if err != nil && !errors.Is(err, ErrTopicNotFound) {
		return nil, errors.Wrap(err, "failed to get current topic")
	}

	// Generate entity key for topic encryption
	entityKey, err := generateSymmetricKey()
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate entity key")
	}

	// Encrypt topic details
	topicDetails := params.GetTopicDetails()
	topicDetailsBytes, err := proto.Marshal(&v1.TopicDetails{
		Title:       topicDetails.Title,
		Description: topicDetails.Description,
		Color:       topicDetails.Color,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal topic details")
	}

	detailsNonce, encryptedDetails, err := encryptWithSymmetricKey(topicDetailsBytes, entityKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to encrypt topic details")
	}

	// Encrypt entity key with diary key
	keyNonce, encryptedEntityKey, err := encryptWithSymmetricKey(entityKey, diaryKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to encrypt entity key")
	}

	encryption := &v1.Encryption{
		DiaryKeyId:        key.Id,
		EncryptedKeyNonce: keyNonce,
		EncryptedKeyData:  encryptedEntityKey,
	}

	if err := validateProtoEncryption(encryption); err != nil {
		return nil, errors.Wrap(err, "request validation failed")
	}

	payload, err := MarshalChangePayload(detailsNonce, encryptedDetails)
	if err != nil {
		return nil, err
	}

	change := &v1.Change{Id: uuid.NewString(), Payload: payload}
	if current != nil {
		if current.GetChange().GetChange() == nil {
			return nil, errors.New("current topic change is required")
		}
		change.ObjectId = topicID
		change.ParentChangeIds = []string{current.GetChange().GetChange().GetId()}
	}

	req := &v1.PutTopicRequest{
		DiaryId:           diaryID,
		TopicId:           topicID,
		DefaultTemplateId: protoString(params.DefaultTemplateID),
		EncryptedChange: &v1.EncryptedChange{
			Change:     change,
			Encryption: encryption,
		},
	}

	response := &v1.PutTopicResponse{}
	if err := c.rpc(ctx, "io.thingsdiary.v1.topic.put", req, response); err != nil {
		return nil, err
	}
	if err := mapTopicErrorCode(response.GetErrorCode()); err != nil {
		return nil, err
	}

	return topicFromProto(response.GetTopic(), diaryKey)
}
