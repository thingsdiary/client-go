package client

import (
	"context"

	v1 "github.com/thingsdiary/client-go/proto"
)

func (c *Client) RestoreEntry(ctx context.Context, diaryID, entryID string) error {
	if c.credentials == nil {
		return ErrUnauthorized
	}

	req := &v1.RestoreEntryRequest{
		DiaryId: diaryID,
		EntryId: entryID,
	}

	// TODO: Sign the encoded request with c.credentials.SigningPrivateKey before sending it.
	if err := c.rpc(ctx, "io.thingsdiary.v1.entry.restore", req, nil); err != nil {
		return err
	}

	return nil
}
