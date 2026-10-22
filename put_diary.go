package client

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"

	v1 "github.com/thingsdiary/client-go/proto"
)

// PutDiaryParams contains parameters for updating a diary
type PutDiaryParams struct {
	Title       string
	Description string
}

// GetDiaryDetails extracts diary details from parameters
func (p PutDiaryParams) GetDiaryDetails() DiaryDetails {
	return DiaryDetails(p)
}

func (c *Client) PutDiary(ctx context.Context, diaryID string, params PutDiaryParams) (*Diary, error) {
	if c.credentials == nil {
		return nil, ErrUnauthorized
	}

	diaryData, err := c.getDiary(ctx, diaryID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get current diary")
	}

	if len(diaryData.EncryptionKeys) == 0 {
		return nil, errors.New("no encryption keys found in diary")
	}

	diaryKeyID := diaryData.EncryptionKeys[0].Id

	entityKey, err := generateSymmetricKey()
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate entity key")
	}

	diaryDetails := params.GetDiaryDetails()
	diaryDetailsJSON, err := json.Marshal(diaryDetails)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal diary details")
	}

	contentNonce, encryptedContent, err := encryptWithSymmetricKey(diaryDetailsJSON, entityKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to encrypt diary content")
	}

	encryptedDiaryKeyValue := diaryData.EncryptionKeys[0].Value
	decryptedDiaryKey, err := decryptWithPrivateKey(
		encryptedDiaryKeyValue,
		c.credentials.EncryptionPrivateKey,
		c.credentials.EncryptionPublicKey,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decrypt diary key")
	}

	keyNonce, encryptedEntityKey, err := encryptWithSymmetricKey(entityKey, decryptedDiaryKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to encrypt entity key")
	}

	encryption := &v1.Encryption{
		DiaryKeyId:        diaryKeyID,
		EncryptedKeyNonce: keyNonce,
		EncryptedKeyData:  encryptedEntityKey,
	}

	details := &v1.EncryptedData{Nonce: contentNonce, Data: encryptedContent}

	if err := validateProtoEncryption(encryption); err != nil {
		return nil, errors.Wrap(err, "request validation failed")
	}

	if err := validateProtoEncryptedData(details, "details"); err != nil {
		return nil, errors.Wrap(err, "request validation failed")
	}

	req := &v1.PutDiaryRequest{
		DiaryId:    diaryID,
		Version:    NewVersion(),
		Encryption: encryption,
		Details:    details,
	}

	response := &v1.Diary{}
	if err := c.rpc(ctx, "io.thingsdiary.v1.diary.put", req, response); err != nil {
		return nil, err
	}

	return diaryFromProto(
		response,
		c.credentials.EncryptionPrivateKey,
		c.credentials.EncryptionPublicKey,
	)
}
