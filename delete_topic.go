package client

import (
	"context"

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

	req := &v1.DeleteTopicRequest{
		DiaryId:       diaryID,
		TopicId:       topicID,
		Soft:          protoBool(p.Soft.OrElse(true)),
		DeleteEntries: p.DeleteEntries,
	}

	if err := c.rpc(ctx, "io.thingsdiary.v1.topic.delete", req, &v1.DeleteTopicResponse{}); err != nil {
		return err
	}

	return nil
}
