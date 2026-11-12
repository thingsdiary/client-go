package client

import (
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/samber/mo"
	v1 "github.com/thingsdiary/client-go/proto"
)

// DeleteTopicParams controls how a topic and its entries are deleted.
type DeleteTopicParams struct {
	DeleteEntries bool
	// Soft defaults to true. Use mo.Some(false) to make the topic permanently removed.
	Soft mo.Option[bool]
}

func (c *Client) DeleteTopic(ctx context.Context, diaryID string, topicID string, params ...DeleteTopicParams) error {
	if c.credentials == nil {
		return ErrUnauthorized
	}

	var p DeleteTopicParams

	if len(params) > 0 {
		p = params[0]
	}
	soft := p.Soft.OrElse(true)

	current, err := c.getTopic(ctx, diaryID, topicID)
	if err != nil {
		if !soft && errors.Is(err, ErrEntityRemoved) {
			return nil
		}

		return err
	}
	encryptedChange, err := nextTopicStateChange(current, topicID)
	if err != nil {
		return err
	}

	req := &v1.DeleteTopicRequest{
		DiaryId:         diaryID,
		TopicId:         topicID,
		Soft:            protoBool(soft),
		DeleteEntries:   p.DeleteEntries,
		EncryptedChange: encryptedChange,
	}

	response := new(v1.DeleteTopicResponse)
	if err := c.rpc(ctx, "io.thingsdiary.v1.topic.delete", req, response); err != nil {
		return err
	}
	if err := mapTopicErrorCode(response.GetErrorCode()); err != nil {
		return err
	}

	return nil
}

func nextTopicStateChange(current *v1.EncryptedTopic, topicID string) (*v1.EncryptedChange, error) {
	if current == nil || current.GetChange().GetChange() == nil || current.GetChange().GetEncryption() == nil {
		return nil, errors.New("current topic change is required")
	}

	return &v1.EncryptedChange{
		Change: &v1.Change{
			Id:              uuid.NewString(),
			ObjectId:        topicID,
			ParentChangeIds: []string{current.GetChange().GetChange().GetId()},
			Payload:         append([]byte(nil), current.GetChange().GetChange().GetPayload()...),
		},
		Encryption: current.GetChange().GetEncryption(),
	}, nil
}
