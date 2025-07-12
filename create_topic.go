package client

import (
	"context"

	"github.com/google/uuid"
	"github.com/samber/mo"
)

type CreateTopicParams struct {
	Title             string
	Description       string
	Color             string
	DefaultTemplateID mo.Option[string]
}

func (c *Client) CreateTopic(ctx context.Context, diaryID string, params CreateTopicParams) (*Topic, error) {
	topicID := uuid.NewString()

	putParams := PutTopicParams(params)

	return c.PutTopic(ctx, diaryID, topicID, putParams)
}
