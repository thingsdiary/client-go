package client

import (
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	v1 "github.com/thingsdiary/client-go/proto"
	"google.golang.org/protobuf/proto"
)

// PutDiaryParams contains parameters for updating a diary
type PutDiaryParams struct {
	Title       string
	Description string
	// ParentChangeIDs identifies the parents of the new change. At least one is required.
	ParentChangeIDs []string
}

// GetDiaryDetails extracts diary details from parameters
func (p PutDiaryParams) GetDiaryDetails() DiaryDetails {
	return DiaryDetails{
		Title:       p.Title,
		Description: p.Description,
	}
}

func (c *Client) PutDiary(ctx context.Context, diaryID string, params PutDiaryParams) (*Diary, error) {
	if c.credentials == nil {
		return nil, ErrUnauthorized
	}

	if len(params.ParentChangeIDs) == 0 {
		return nil, errors.Wrap(ErrBadRequest, "parent change IDs are required")
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
	diaryDetailsBytes, err := proto.Marshal(&v1.DiaryDetails{
		Title:       diaryDetails.Title,
		Description: diaryDetails.Description,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal diary details")
	}

	contentNonce, encryptedContent, err := encryptWithSymmetricKey(diaryDetailsBytes, entityKey)
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

	if err := validateProtoEncryption(encryption); err != nil {
		return nil, errors.Wrap(err, "request validation failed")
	}

	payload, err := MarshalChangePayload(contentNonce, encryptedContent)
	if err != nil {
		return nil, err
	}

	change := &v1.Change{
		Id:              uuid.NewString(),
		ObjectId:        diaryID,
		ParentChangeIds: append([]string(nil), params.ParentChangeIDs...),
		Payload:         payload,
	}

	req := &v1.PutDiaryRequest{
		DiaryId: diaryID,
		EncryptedChange: &v1.EncryptedChange{
			Change:     change,
			Encryption: encryption,
		},
	}

	response := &v1.PutDiaryResponse{}
	if err := c.rpc(ctx, "io.thingsdiary.v1.diary.put", req, response); err != nil {
		return nil, err
	}

	diaryData, err = c.getDiary(ctx, diaryID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get updated diary")
	}
	diary, err := diaryFromProto(
		diaryData,
		c.credentials.EncryptionPrivateKey,
		c.credentials.EncryptionPublicKey,
	)
	if err != nil {
		return nil, err
	}

	return diary, nil
}
