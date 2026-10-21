package client

import (
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	v1 "github.com/thingsdiary/client-go/proto"
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

func diaryFromProto(value *v1.Diary, privateKey, publicKey []byte) (*Diary, error) {
	if value == nil {
		return nil, errors.New("diary is required")
	}

	if len(value.EncryptionKeys) == 0 {
		return nil, errors.New("no encryption keys found in response")
	}

	decryptedDiaryKey, err := decryptWithPrivateKey(
		value.EncryptionKeys[0].Value,
		privateKey,
		publicKey,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decrypt diary key")
	}

	decryptedContentBytes, err := decryptProtoDetails(
		value.Encryption,
		value.Details,
		decryptedDiaryKey,
		"diary content",
	)
	if err != nil {
		return nil, err
	}

	var decryptedDiaryDetails DiaryDetails
	if err := json.Unmarshal(decryptedContentBytes, &decryptedDiaryDetails); err != nil {
		return nil, errors.Wrap(err, "failed to parse decrypted diary details")
	}

	diary := &Diary{
		ID:          value.Id,
		Title:       decryptedDiaryDetails.Title,
		Description: decryptedDiaryDetails.Description,
		CreatedAt:   protoTime(value.CreatedAt),
		UpdatedAt:   protoTime(value.UpdatedAt),
		Version:     value.Version,
	}

	return diary, nil
}
