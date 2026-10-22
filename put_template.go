package client

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"

	v1 "github.com/thingsdiary/client-go/proto"
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
	templateDetailsJSON, err := json.Marshal(templateDetails)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal template details")
	}

	detailsNonce, encryptedDetails, err := encryptWithSymmetricKey(templateDetailsJSON, entityKey)
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

	details := &v1.EncryptedData{Nonce: detailsNonce, Data: encryptedDetails}

	if err := validateProtoEncryption(encryption); err != nil {
		return nil, errors.Wrap(err, "request validation failed")
	}

	if err := validateProtoEncryptedData(details, "details"); err != nil {
		return nil, errors.Wrap(err, "request validation failed")
	}

	req := &v1.PutTemplateRequest{
		DiaryId:    diaryID,
		TemplateId: templateID,
		Version:    NewVersion(),
		Encryption: encryption,
		Details:    details,
	}

	response := &v1.Template{}
	if err := c.rpc(ctx, "io.thingsdiary.v1.template.put", req, response); err != nil {
		return nil, err
	}

	return templateFromProto(response, decryptedDiaryKey)
}
