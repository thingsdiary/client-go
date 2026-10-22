package client

import (
	"context"

	"github.com/samber/mo"
	v1 "github.com/thingsdiary/client-go/proto"
)

// DeleteEntryParams controls how an entry is deleted.
type DeleteEntryParams struct {
	// Soft defaults to true. Use mo.Some(false) to make the entry permanently removed.
	Soft mo.Option[bool]
}

func (c *Client) DeleteEntry(ctx context.Context, diaryID string, entryID string, params ...DeleteEntryParams) error {
	if c.credentials == nil {
		return ErrUnauthorized
	}

	var p DeleteEntryParams

	if len(params) > 0 {
		p = params[0]
	}

	req := &v1.DeleteEntryRequest{
		DiaryId: diaryID,
		EntryId: entryID,
		Soft:    protoBool(p.Soft.OrElse(true)),
	}

	if err := c.rpc(ctx, "io.thingsdiary.v1.entry.delete", req, &v1.DeleteEntryResponse{}); err != nil {
		return err
	}

	return nil
}
