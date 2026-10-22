package client

import (
	"context"

	"github.com/pkg/errors"
	v1 "github.com/thingsdiary/client-go/proto"
)

func (c *Client) GetTemplateByID(ctx context.Context, diaryID, templateID string) (*Template, error) {
	if c.credentials == nil {
		return nil, ErrUnauthorized
	}

	// Get encryption keys
	key, err := c.getActiveDiaryKey(ctx, diaryID)
	if err != nil {
		return nil, err
	}

	// Decrypt diary key
	decryptedDiaryKey, err := decryptWithPrivateKey(
		key.Value,
		c.credentials.EncryptionPrivateKey,
		c.credentials.EncryptionPublicKey,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decrypt diary key")
	}

	value, err := c.getTemplate(ctx, diaryID, templateID)
	if err != nil {
		return nil, err
	}

	return templateFromProto(value, decryptedDiaryKey)
}

func (c *Client) getTemplate(ctx context.Context, diaryID, templateID string) (*v1.Template, error) {
	req := &v1.GetTemplateRequest{
		DiaryId:    diaryID,
		TemplateId: templateID,
	}

	response := &v1.GetTemplateResponse{}

	if err := c.rpc(ctx, "io.thingsdiary.v1.template.get", req, response); err != nil {
		return nil, err
	}

	return response.Template, nil
}
