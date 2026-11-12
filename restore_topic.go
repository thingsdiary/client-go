package client

import (
	"context"

	"github.com/pkg/errors"
	v1 "github.com/thingsdiary/client-go/proto"
)

func (c *Client) RestoreTopic(ctx context.Context, diaryID, topicID string) (*Topic, error) {
	if c.credentials == nil {
		return nil, ErrUnauthorized
	}

	current, err := c.getTopic(ctx, diaryID, topicID)
	if err != nil {
		return nil, err
	}
	encryptedChange, err := nextTopicStateChange(current, topicID)
	if err != nil {
		return nil, err
	}

	req := &v1.RestoreTopicRequest{
		DiaryId:         diaryID,
		TopicId:         topicID,
		EncryptedChange: encryptedChange,
	}

	response := &v1.RestoreTopicResponse{}
	if err := c.rpc(ctx, "io.thingsdiary.v1.topic.restore", req, response); err != nil {
		return nil, err
	}
	if err := mapTopicErrorCode(response.GetErrorCode()); err != nil {
		return nil, err
	}

	keys, err := c.getDiaryKeys(ctx, diaryID)
	if err != nil {
		return nil, err
	}
	key, err := diaryKeyByID(keys, response.GetTopic().GetChange().GetEncryption().GetDiaryKeyId())
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

	return topicFromProto(response.GetTopic(), diaryKey)
}
