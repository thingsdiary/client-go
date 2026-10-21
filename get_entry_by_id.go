package client

import (
	"context"

	"github.com/pkg/errors"
	v1 "github.com/thingsdiary/client-go/proto"
)

func (c *Client) GetEntryByID(ctx context.Context, diaryID, entryID string) (*Entry, error) {
	if c.credentials == nil {
		return nil, ErrUnauthorized
	}

	// Get encryption keys
	key, err := c.getActiveDiaryKey(ctx, diaryID)
	if err != nil {
		return nil, err
	}

	// Decrypt diary key
	decryptedDiaryKey, err := decryptWithPrivateKey(
		key.Value,
		c.credentials.EncryptionPrivateKey,
		c.credentials.EncryptionPublicKey,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decrypt diary key")
	}

	// Get entry data
	value, err := c.getEntry(ctx, diaryID, entryID)
	if err != nil {
		return nil, err
	}

	// Decrypt and return entry
	return entryFromProto(value, decryptedDiaryKey)
}

func (c *Client) getEntry(ctx context.Context, diaryID, entryID string) (*v1.Entry, error) {
	req := &v1.GetEntryRequest{
		DiaryId: diaryID,
		EntryId: entryID,
	}

	response := &v1.GetEntryResponse{}

	if err := c.rpc(ctx, "io.thingsdiary.v1.entry.get", req, response); err != nil {
		return nil, err
	}

	return response.Entry, nil
}
