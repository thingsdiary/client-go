package client

import (
	"context"

	v1 "github.com/thingsdiary/client-go/proto"
)

func (c *Client) RestoreTopic(ctx context.Context, diaryID, topicID string) error {
	if c.credentials == nil {
		return ErrUnauthorized
	}

	req := &v1.RestoreTopicRequest{
		DiaryId: diaryID,
		TopicId: topicID,
	}

	if err := c.rpc(ctx, "io.thingsdiary.v1.topic.restore", req, nil); err != nil {
		return err
	}

	return nil
}
