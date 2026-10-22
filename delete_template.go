package client

import (
	"context"

	"github.com/samber/mo"
	v1 "github.com/thingsdiary/client-go/proto"
)

// DeleteTemplateParams controls how a template is deleted.
type DeleteTemplateParams struct {
	// Soft defaults to true. Use mo.Some(false) to make the template permanently removed.
	Soft mo.Option[bool]
}

func (c *Client) DeleteTemplate(ctx context.Context, diaryID string, templateID string, params ...DeleteTemplateParams) error {
	if c.credentials == nil {
		return ErrUnauthorized
	}

	var p DeleteTemplateParams

	if len(params) > 0 {
		p = params[0]
	}

	req := &v1.DeleteTemplateRequest{
		DiaryId:    diaryID,
		TemplateId: templateID,
		Soft:       protoBool(p.Soft.OrElse(true)),
	}

	if err := c.rpc(ctx, "io.thingsdiary.v1.template.delete", req, &v1.DeleteTemplateResponse{}); err != nil {
		return err
	}

	return nil
}
