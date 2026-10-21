package client

import (
	"context"

	v1 "github.com/thingsdiary/client-go/proto"
)

type DeleteEntryParams struct {
	Soft *bool
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
		Soft:    protoBool(deleteSoftValue(p.Soft)),
	}

	// TODO: Sign the encoded request with c.credentials.SigningPrivateKey before sending it.
	if err := c.rpc(ctx, "io.thingsdiary.v1.entry.delete", req, nil); err != nil {
		return err
	}

	return nil
}
