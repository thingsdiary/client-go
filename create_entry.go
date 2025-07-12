package client

import (
	"context"

	"github.com/google/uuid"
	"github.com/samber/mo"
)

type CreateEntryParams struct {
	Content       string
	TopicID       mo.Option[string]
	Archived      bool
	Bookmarked    bool
	PreviewHidden bool
}

func (c *Client) CreateEntry(ctx context.Context, diaryID string, params CreateEntryParams) (*Entry, error) {
	entryID := uuid.NewString()

	putParams := PutEntryParams(params)

	return c.PutEntry(ctx, diaryID, entryID, putParams)
}
