package client

import (
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	v1 "github.com/thingsdiary/client-go/proto"
	"google.golang.org/protobuf/proto"
)

// CreateDiaryParams contains parameters for creating a new diary
type CreateDiaryParams struct {
	Title       string
	Description string
}

// GetDiaryDetails extracts diary details from parameters
func (p CreateDiaryParams) GetDiaryDetails() DiaryDetails {
	return DiaryDetails(p)
}

// CreateDiary creates a new diary with zero-knowledge encryption
func (c *Client) CreateDiary(ctx context.Context, params CreateDiaryParams) (*Diary, error) {
	if c.credentials == nil {
		return nil, ErrUnauthorized
	}

	// Generate diary key for envelope encryption
	diaryKey, err := generateSymmetricKey()
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate diary key")
	}

	// Generate entity key for content encryption
	entityKey, err := generateSymmetricKey()
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate entity key")
	}

	// Encrypt diary content with entity key
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

	// Encrypt entity key with diary key
	keyNonce, encryptedEntityKey, err := encryptWithSymmetricKey(entityKey, diaryKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to encrypt entity key")
	}

	// Encrypt diary key with user's public key (envelope encryption)
	encryptedDiaryKey, err := encryptWithPublicKey(diaryKey, c.credentials.EncryptionPublicKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to encrypt diary key")
	}

	payload, err := MarshalChangePayload(contentNonce, encryptedContent)
	if err != nil {
		return nil, err
	}

	change := &v1.Change{
		Id:      uuid.NewString(),
		Payload: payload,
	}

	req := &v1.CreateDiaryRequest{
		Change:            change,
		EncryptedDiaryKey: encryptedDiaryKey,
		EncryptedKeyNonce: keyNonce,
		EncryptedKeyData:  encryptedEntityKey,
		Details: &v1.EncryptedData{
			Nonce: contentNonce,
			Data:  encryptedContent,
		},
	}

	response := &v1.CreateDiaryResponse{}
	if err := c.rpc(ctx, "io.thingsdiary.v1.diary.create", req, response); err != nil {
		return nil, err
	}

	diary, err := diaryFromProto(
		response.GetDiary(),
		c.credentials.EncryptionPrivateKey,
		c.credentials.EncryptionPublicKey,
	)
	if err != nil {
		return nil, err
	}

	diary.ChangeID = change.GetId()

	return diary, nil
}
