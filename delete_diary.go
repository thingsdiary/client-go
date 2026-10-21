package client

import (
	"context"

	v1 "github.com/thingsdiary/client-go/proto"
)

func (c *Client) DeleteDiary(ctx context.Context, diaryID string) error {
	if c.credentials == nil {
		return ErrUnauthorized
	}

	req := &v1.DeleteDiaryRequest{
		DiaryId: diaryID,
	}

	if err := c.rpc(ctx, "io.thingsdiary.v1.diary.delete", req, &v1.DeleteDiaryResponse{}); err != nil {
		return err
	}

	return nil
}
