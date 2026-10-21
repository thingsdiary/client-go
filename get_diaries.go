package client

import (
	"context"

	v1 "github.com/thingsdiary/client-go/proto"
)

func (c *Client) GetDiaries(ctx context.Context) ([]*Diary, error) {
	if c.credentials == nil {
		return nil, ErrUnauthorized
	}

	response, err := c.getDiaries(ctx)
	if err != nil {
		return nil, err
	}

	diaries := make([]*Diary, 0, len(response.Diaries))
	for _, value := range response.Diaries {
		diary, err := diaryFromProto(
			value,
			c.credentials.EncryptionPrivateKey,
			c.credentials.EncryptionPublicKey,
		)
		if err != nil {
			return nil, err
		}
		diaries = append(diaries, diary)
	}

	return diaries, nil
}

func (c *Client) getDiaries(ctx context.Context) (*v1.ListDiariesResponse, error) {
	req := &v1.ListDiariesRequest{}

	response := &v1.ListDiariesResponse{}

	if err := c.rpc(ctx, "io.thingsdiary.v1.diary.list", req, response); err != nil {
		return nil, err
	}

	return response, nil
}
