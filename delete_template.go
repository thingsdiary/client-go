package client

import (
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"
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
	soft := p.Soft.OrElse(true)

	current, err := c.getTemplate(ctx, diaryID, templateID)
	if err != nil {
		if !soft && errors.Is(err, ErrEntityRemoved) {
			return nil
		}

		return err
	}
	encryptedChange, err := nextTemplateStateChange(current, templateID)
	if err != nil {
		return err
	}

	req := &v1.DeleteTemplateRequest{
		DiaryId:         diaryID,
		TemplateId:      templateID,
		Soft:            protoBool(soft),
		EncryptedChange: encryptedChange,
	}

	response := new(v1.DeleteTemplateResponse)
	if err := c.rpc(ctx, "io.thingsdiary.v1.template.delete", req, response); err != nil {
		return err
	}
	if err := mapTemplateErrorCode(response.GetErrorCode()); err != nil {
		return err
	}

	return nil
}

func nextTemplateStateChange(current *v1.EncryptedTemplate, templateID string) (*v1.EncryptedChange, error) {
	if current == nil || current.GetChange().GetChange() == nil || current.GetChange().GetEncryption() == nil {
		return nil, errors.New("current template change is required")
	}

	return &v1.EncryptedChange{
		Change: &v1.Change{
			Id:              uuid.NewString(),
			ObjectId:        templateID,
			ParentChangeIds: []string{current.GetChange().GetChange().GetId()},
			Payload:         append([]byte(nil), current.GetChange().GetChange().GetPayload()...),
		},
		Encryption: current.GetChange().GetEncryption(),
	}, nil
}
