package client

import (
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	v1 "github.com/thingsdiary/client-go/proto"
	"google.golang.org/protobuf/proto"
)

// PutTemplateParams contains parameters for creating/updating a template
type PutTemplateParams struct {
	Content string
}

// GetTemplateDetails extracts template details from parameters
func (p PutTemplateParams) GetTemplateDetails() TemplateDetails {
	return TemplateDetails(p)
}

// PutTemplate creates or updates a template in a diary
func (c *Client) PutTemplate(ctx context.Context, diaryID, templateID string, params PutTemplateParams) (*Template, error) {
	if c.credentials == nil {
		return nil, ErrUnauthorized
	}

	// Get encryption keys
	key, err := c.getActiveDiaryKey(ctx, diaryID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get active diary key")
	}

	diaryKeyID := key.Id

	// Generate entity key for template encryption
	entityKey, err := generateSymmetricKey()
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate entity key")
	}

	// Encrypt template details
	templateDetails := params.GetTemplateDetails()
	templateDetailsBytes, err := proto.Marshal(&v1.TemplateDetails{
		Content: templateDetails.Content,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal template details")
	}

	detailsNonce, encryptedDetails, err := encryptWithSymmetricKey(templateDetailsBytes, entityKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to encrypt template details")
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

	current, err := c.getTemplate(ctx, diaryID, templateID)
	if err != nil && !errors.Is(err, ErrTemplateNotFound) {
		return nil, errors.Wrap(err, "failed to get current template")
	}

	// Encrypt entity key with diary key
	keyNonce, encryptedEntityKey, err := encryptWithSymmetricKey(entityKey, decryptedDiaryKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to encrypt entity key")
	}

	encryption := &v1.Encryption{
		DiaryKeyId:        diaryKeyID,
		EncryptedKeyNonce: keyNonce,
		EncryptedKeyData:  encryptedEntityKey,
	}

	if err := validateProtoEncryption(encryption); err != nil {
		return nil, errors.Wrap(err, "request validation failed")
	}

	payload, err := MarshalChangePayload(detailsNonce, encryptedDetails)
	if err != nil {
		return nil, err
	}

	change := &v1.Change{Id: uuid.NewString(), Payload: payload}
	if current != nil {
		if current.GetChange().GetChange() == nil {
			return nil, errors.New("current template change is required")
		}
		change.ObjectId = templateID
		change.ParentChangeIds = []string{current.GetChange().GetChange().GetId()}
	}

	req := &v1.PutTemplateRequest{
		DiaryId:    diaryID,
		TemplateId: templateID,
		EncryptedChange: &v1.EncryptedChange{
			Change:     change,
			Encryption: encryption,
		},
	}

	response := &v1.PutTemplateResponse{}
	if err := c.rpc(ctx, "io.thingsdiary.v1.template.put", req, response); err != nil {
		return nil, err
	}
	if err := mapTemplateErrorCode(response.GetErrorCode()); err != nil {
		return nil, err
	}

	return templateFromProto(response.GetTemplate(), decryptedDiaryKey)
}
