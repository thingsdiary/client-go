package client

import (
	"context"

	"github.com/pkg/errors"
	v1 "github.com/thingsdiary/client-go/proto"
)

func (c *Client) GetTemplates(ctx context.Context, diaryID string) ([]*Template, error) {
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

	// Get templates data
	response, err := c.getTemplates(ctx, diaryID)
	if err != nil {
		return nil, err
	}

	templates := make([]*Template, 0, len(response.Templates))
	for _, value := range response.Templates {
		template, err := templateFromProto(value, decryptedDiaryKey)
		if err != nil {
			return nil, err
		}

		templates = append(templates, template)
	}

	return templates, nil
}

func (c *Client) getTemplates(ctx context.Context, diaryID string) (*v1.ListTemplatesResponse, error) {
	req := &v1.ListTemplatesRequest{
		DiaryId: diaryID,
	}

	response := &v1.ListTemplatesResponse{}

	if err := c.rpc(ctx, "io.thingsdiary.v1.template.list", req, response); err != nil {
		return nil, err
	}

	return response, nil
}
