package client

import (
	"context"

	"github.com/pkg/errors"

	v1 "github.com/thingsdiary/client-go/proto"
)

func (c *Client) getActiveDiaryKey(ctx context.Context, diaryID string) (*v1.DiaryEncryptionKey, error) {
	// TODO: Implement client-side caching for keys (TTL ~5min)
	req := &v1.ListDiaryKeysRequest{
		DiaryId: diaryID,
	}

	response := &v1.ListDiaryKeysResponse{}

	if err := c.rpc(ctx, "io.thingsdiary.v1.diary_key.list", req, response); err != nil {
		return nil, err
	}

	for _, key := range response.Keys {
		if key != nil && key.Status == v1.DiaryKeyStatus_DIARY_KEY_STATUS_ACTIVE {
			return key, nil
		}
	}

	return nil, errors.New("no active encryption key found")
}
