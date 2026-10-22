package client

import (
	"context"

	v1 "github.com/thingsdiary/client-go/proto"
)

func (c *Client) RestoreTemplate(ctx context.Context, diaryID, templateID string) error {
	if c.credentials == nil {
		return ErrUnauthorized
	}

	req := &v1.RestoreTemplateRequest{
		DiaryId:    diaryID,
		TemplateId: templateID,
	}

	if err := c.rpc(ctx, "io.thingsdiary.v1.template.restore", req, nil); err != nil {
		return err
	}

	return nil
}
