package client

import (
	"context"

	v1 "github.com/thingsdiary/client-go/proto"
)

type DeleteTemplateParams struct {
	Soft *bool
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
		Soft:       protoBool(deleteSoftValue(p.Soft)),
	}

	// TODO: Sign the encoded request with c.credentials.SigningPrivateKey before sending it.
	if err := c.rpc(ctx, "io.thingsdiary.v1.template.delete", req, nil); err != nil {
		return err
	}

	return nil
}
