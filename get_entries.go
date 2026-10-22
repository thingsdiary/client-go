package client

import (
	"context"

	"github.com/pkg/errors"
	v1 "github.com/thingsdiary/client-go/proto"
)

func (c *Client) GetEntries(ctx context.Context, diaryID string) ([]*Entry, error) {
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

	// Get entries data
	response, err := c.getEntries(ctx, diaryID)
	if err != nil {
		return nil, err
	}

	entries := make([]*Entry, 0, len(response.Entries))
	for _, value := range response.Entries {
		entry, err := entryFromProto(value, decryptedDiaryKey)
		if err != nil {
			return nil, err
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

func (c *Client) getEntries(ctx context.Context, diaryID string) (*v1.ListEntriesResponse, error) {
	req := &v1.ListEntriesRequest{
		DiaryId: diaryID,
	}

	response := &v1.ListEntriesResponse{}

	if err := c.rpc(ctx, "io.thingsdiary.v1.entry.list", req, response); err != nil {
		return nil, err
	}

	return response, nil
}
