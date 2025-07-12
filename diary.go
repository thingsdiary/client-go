package client

import (
	"encoding/json"
	"time"

	"github.com/pkg/errors"

	"github.com/thingsdiary/client-go/openapi"
)

// Diary represents a plaintext diary that users work with
type Diary struct {
	ID          string
	Title       string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Version     uint64
}

// DiaryDetails represents the plaintext content structure for diary details
type DiaryDetails struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// decryptDiary decrypts a diary using the provided credentials
func (c *Client) decryptDiary(diaryData *openapi.Diary) (*Diary, error) {
	if c.credentials == nil {
		return nil, ErrUnauthorized
	}

	if len(diaryData.EncryptionKeys) == 0 {
		return nil, errors.New("no encryption keys found in response")
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

	decryptedEntityKey, err := decryptWithSymmetricKey(
		diaryData.Encryption.EncryptedKeyNonce,
		diaryData.Encryption.EncryptedKeyData,
		decryptedDiaryKey,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decrypt entity key")
	}

	decryptedContentBytes, err := decryptWithSymmetricKey(
		diaryData.Details.Nonce,
		diaryData.Details.Data,
		decryptedEntityKey,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decrypt diary content")
	}

	var decryptedDiaryDetails DiaryDetails
	if err := json.Unmarshal(decryptedContentBytes, &decryptedDiaryDetails); err != nil {
		return nil, errors.Wrap(err, "failed to parse decrypted diary details")
	}

	diary := &Diary{
		ID:          diaryData.Id,
		Title:       decryptedDiaryDetails.Title,
		Description: decryptedDiaryDetails.Description,
		CreatedAt:   diaryData.CreatedAt,
		UpdatedAt:   diaryData.UpdatedAt,
		Version:     diaryData.Version,
	}

	return diary, nil
}
