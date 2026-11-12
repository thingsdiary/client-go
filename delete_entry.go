package client

import (
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"
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
	soft := p.Soft.OrElse(true)

	current, err := c.getEntry(ctx, diaryID, entryID)
	if err != nil {
		if !soft && errors.Is(err, ErrEntityRemoved) {
			return nil
		}

		return err
	}
	encryptedChange, err := nextEntryStateChange(current, entryID)
	if err != nil {
		return err
	}

	req := &v1.DeleteEntryRequest{
		DiaryId:         diaryID,
		EntryId:         entryID,
		Soft:            protoBool(soft),
		EncryptedChange: encryptedChange,
	}

	response := new(v1.DeleteEntryResponse)
	if err := c.rpc(ctx, "io.thingsdiary.v1.entry.delete", req, response); err != nil {
		return err
	}
	if err := mapEntryErrorCode(response.GetErrorCode()); err != nil {
		return err
	}

	return nil
}

func nextEntryStateChange(current *v1.EncryptedEntry, entryID string) (*v1.EncryptedChange, error) {
	if current == nil || current.GetChange().GetChange() == nil || current.GetChange().GetEncryption() == nil {
		return nil, errors.New("current entry change is required")
	}

	return &v1.EncryptedChange{
		Change: &v1.Change{
			Id:              uuid.NewString(),
			ObjectId:        entryID,
			ParentChangeIds: []string{current.GetChange().GetChange().GetId()},
			Payload:         append([]byte(nil), current.GetChange().GetChange().GetPayload()...),
		},
		Encryption: current.GetChange().GetEncryption(),
	}, nil
}
