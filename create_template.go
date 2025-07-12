package client

import (
	"context"

	"github.com/google/uuid"
)

type CreateTemplateParams struct {
	Content string
}

func (c *Client) CreateTemplate(ctx context.Context, diaryID string, params CreateTemplateParams) (*Template, error) {
	templateID := uuid.NewString()

	putParams := PutTemplateParams(params)

	return c.PutTemplate(ctx, diaryID, templateID, putParams)
}
