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

	response := &v1.DeleteDiaryResponse{}
	if err := c.rpc(ctx, "io.thingsdiary.v1.diary.delete", req, response); err != nil {
		return err
	}
	if err := mapDiaryErrorCode(response.GetErrorCode()); err != nil {
		return err
	}

	return nil
}
