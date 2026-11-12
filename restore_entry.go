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

	current, err := c.getEntry(ctx, diaryID, entryID)
	if err != nil {
		return nil, err
	}
	encryptedChange, err := nextEntryStateChange(current, entryID)
	if err != nil {
		return nil, err
	}

	req := &v1.RestoreEntryRequest{
		DiaryId:         diaryID,
		EntryId:         entryID,
		EncryptedChange: encryptedChange,
	}

	response := &v1.RestoreEntryResponse{}
	if err := c.rpc(ctx, "io.thingsdiary.v1.entry.restore", req, response); err != nil {
		return nil, err
	}
	if err := mapEntryErrorCode(response.GetErrorCode()); err != nil {
		return nil, err
	}

	keys, err := c.getDiaryKeys(ctx, diaryID)
	if err != nil {
		return nil, err
	}
	key, err := diaryKeyByID(keys, response.GetEntry().GetChange().GetEncryption().GetDiaryKeyId())
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
