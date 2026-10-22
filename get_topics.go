package client

import (
	"context"

	"github.com/pkg/errors"
	v1 "github.com/thingsdiary/client-go/proto"
)

func (c *Client) GetTopics(ctx context.Context, diaryID string) ([]*Topic, error) {
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

	// Get topics data
	response, err := c.getTopics(ctx, diaryID)
	if err != nil {
		return nil, err
	}

	topics := make([]*Topic, 0, len(response.Topics))
	for _, value := range response.Topics {
		topic, err := topicFromProto(value, decryptedDiaryKey)
		if err != nil {
			return nil, err
		}

		topics = append(topics, topic)
	}

	return topics, nil
}

func (c *Client) getTopics(ctx context.Context, diaryID string) (*v1.ListTopicsResponse, error) {
	req := &v1.ListTopicsRequest{
		DiaryId: diaryID,
	}

	response := &v1.ListTopicsResponse{}

	if err := c.rpc(ctx, "io.thingsdiary.v1.topic.list", req, response); err != nil {
		return nil, err
	}

	return response, nil
}
