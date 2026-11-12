package client

import (
	"time"

	"github.com/pkg/errors"
	v1 "github.com/thingsdiary/client-go/proto"
	"google.golang.org/protobuf/proto"
)

// Diary represents a plaintext diary that users work with
type Diary struct {
	ID          string
	Title       string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	// Version is a local compatibility value derived from UpdatedAt. Diary
	// versions are no longer part of the encrypted DAG contract.
	Version uint64
	// ChangeID identifies the change returned by the diary read model.
	ChangeID string
}

// DiaryDetails represents the plaintext content structure for diary details
type DiaryDetails struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

func diaryFromProto(value *v1.EncryptedDiary, privateKey, publicKey []byte) (*Diary, error) {
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

	change := value.GetChange()
	if change == nil || change.GetChange() == nil || change.GetEncryption() == nil {
		return nil, errors.New("diary change is required")
	}

	nonce, encryptedContent, err := UnmarshalChangePayload(change.GetChange().GetPayload())
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode diary change payload")
	}
	entityKey, err := decryptWithSymmetricKey(
		change.GetEncryption().GetEncryptedKeyNonce(),
		change.GetEncryption().GetEncryptedKeyData(),
		decryptedDiaryKey,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decrypt diary entity key")
	}
	decryptedContentBytes, err := decryptWithSymmetricKey(nonce, encryptedContent, entityKey)
	if err != nil {
		return nil, err
	}

	var decryptedDiaryDetails v1.DiaryDetails
	if err := proto.Unmarshal(decryptedContentBytes, &decryptedDiaryDetails); err != nil {
		return nil, errors.Wrap(err, "failed to parse decrypted diary details")
	}

	diary := &Diary{
		ID:          value.Id,
		Title:       decryptedDiaryDetails.Title,
		Description: decryptedDiaryDetails.Description,
		CreatedAt:   protoTime(value.CreatedAt),
		UpdatedAt:   protoTime(value.UpdatedAt),
		Version:     uint64(protoTime(value.UpdatedAt).UnixNano()),
		ChangeID:    change.GetChange().GetId(),
	}

	return diary, nil
}
