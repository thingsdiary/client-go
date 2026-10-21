package client

import (
	"context"

	v1 "github.com/thingsdiary/client-go/proto"
)

func (c *Client) GetDiaryByID(ctx context.Context, id string) (*Diary, error) {
	if c.credentials == nil {
		return nil, ErrUnauthorized
	}

	diary, err := c.getDiary(ctx, id)
	if err != nil {
		return nil, err
	}

	return diaryFromProto(
		diary,
		c.credentials.EncryptionPrivateKey,
		c.credentials.EncryptionPublicKey,
	)
}

func (c *Client) getDiary(ctx context.Context, id string) (*v1.Diary, error) {
	req := &v1.GetDiaryRequest{
		DiaryId: id,
	}

	response := &v1.GetDiaryResponse{}

	if err := c.rpc(ctx, "io.thingsdiary.v1.diary.get", req, response); err != nil {
		return nil, err
	}

	return response.Diary, nil
}
