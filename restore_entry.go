package client

import (
	"context"

	"github.com/pkg/errors"

	v1 "github.com/thingsdiary/client-go/proto"
)

func (c *Client) RestoreEntry(ctx context.Context, diaryID, entryID string) (*Entry, error) {
	if c.credentials == nil {
		return nil, ErrUnauthorized
	}

	keys, err := c.getDiaryKeys(ctx, diaryID)
	if err != nil {
		return nil, err
	}

	req := &v1.RestoreEntryRequest{
		DiaryId: diaryID,
		EntryId: entryID,
	}

	response := &v1.RestoreEntryResponse{}
	if err := c.rpc(ctx, "io.thingsdiary.v1.entry.restore", req, response); err != nil {
		return nil, err
	}

	key, err := diaryKeyByID(keys, response.GetEntry().GetEncryption().GetDiaryKeyId())
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

	return entryFromProto(response.GetEntry(), diaryKey)
}
