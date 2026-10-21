package client

import (
	"context"

	"github.com/pkg/errors"

	v1 "github.com/thingsdiary/client-go/proto"
)

func (c *Client) GetTopicByID(ctx context.Context, diaryID, topicID string) (*Topic, error) {
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

	// Get topic data
	value, err := c.getTopic(ctx, diaryID, topicID)
	if err != nil {
		return nil, err
	}

	return topicFromProto(value, decryptedDiaryKey)
}

func (c *Client) getTopic(ctx context.Context, diaryID, topicID string) (*v1.Topic, error) {
	req := &v1.GetTopicRequest{
		DiaryId: diaryID,
		TopicId: topicID,
	}

	response := &v1.GetTopicResponse{}

	if err := c.rpc(ctx, "io.thingsdiary.v1.topic.get", req, response); err != nil {
		return nil, err
	}

	return response.Topic, nil
}
