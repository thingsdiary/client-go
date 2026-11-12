package client

import (
	"context"

	"github.com/pkg/errors"
	v1 "github.com/thingsdiary/client-go/proto"
)

func (c *Client) RestoreTemplate(ctx context.Context, diaryID, templateID string) (*Template, error) {
	if c.credentials == nil {
		return nil, ErrUnauthorized
	}

	current, err := c.getTemplate(ctx, diaryID, templateID)
	if err != nil {
		return nil, err
	}
	encryptedChange, err := nextTemplateStateChange(current, templateID)
	if err != nil {
		return nil, err
	}

	req := &v1.RestoreTemplateRequest{
		DiaryId:         diaryID,
		TemplateId:      templateID,
		EncryptedChange: encryptedChange,
	}

	response := &v1.RestoreTemplateResponse{}
	if err := c.rpc(ctx, "io.thingsdiary.v1.template.restore", req, response); err != nil {
		return nil, err
	}
	if err := mapTemplateErrorCode(response.GetErrorCode()); err != nil {
		return nil, err
	}

	keys, err := c.getDiaryKeys(ctx, diaryID)
	if err != nil {
		return nil, err
	}
	key, err := diaryKeyByID(keys, response.GetTemplate().GetChange().GetEncryption().GetDiaryKeyId())
	if err != nil {
		return nil, err
	}
	diaryKey, err := decryptWithPrivateKey(
		key.Value,
		c.credentials.EncryptionPrivateKey,
		c.credentials.EncryptionPublicKey,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decrypt diary key")
	}

	return templateFromProto(response.GetTemplate(), diaryKey)
}
