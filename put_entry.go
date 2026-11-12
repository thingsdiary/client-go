package client

import (
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/samber/mo"
	v1 "github.com/thingsdiary/client-go/proto"
	"google.golang.org/protobuf/proto"
)

// PutEntryParams contains parameters for creating/updating an entry
type PutEntryParams struct {
	Content       string
	TopicID       mo.Option[string]
	Archived      bool
	Bookmarked    bool
	PreviewHidden bool
}

// GetEntryDetails extracts entry details from parameters
func (p PutEntryParams) GetEntryDetails() EntryDetails {
	return EntryDetails{
		Content:       p.Content,
		Archived:      p.Archived,
		Bookmarked:    p.Bookmarked,
		PreviewHidden: p.PreviewHidden,
	}
}

// PutEntry creates or updates an entry in a diary
func (c *Client) PutEntry(ctx context.Context, diaryID, entryID string, params PutEntryParams) (*Entry, error) {
	if c.credentials == nil {
		return nil, ErrUnauthorized
	}

	// Get encryption keys
	key, err := c.getActiveDiaryKey(ctx, diaryID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get active diary key")
	}

	diaryKeyID := key.Id
	diaryKey, err := decryptWithPrivateKey(
		key.Value,
		c.credentials.EncryptionPrivateKey,
		c.credentials.EncryptionPublicKey,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decrypt diary key")
	}

	current, err := c.getEntry(ctx, diaryID, entryID)
	if err != nil && !errors.Is(err, ErrEntryNotFound) {
		return nil, errors.Wrap(err, "failed to get current entry")
	}

	// Generate entity key for entry encryption
	entityKey, err := generateSymmetricKey()
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate entity key")
	}

	// Encrypt entry details
	entryDetails := params.GetEntryDetails()
	entryDetailsBytes, err := proto.Marshal(&v1.EntryDetails{
		Content:       entryDetails.Content,
		Archived:      entryDetails.Archived,
		Bookmarked:    entryDetails.Bookmarked,
		PreviewHidden: entryDetails.PreviewHidden,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal entry details")
	}

	detailsNonce, encryptedDetails, err := encryptWithSymmetricKey(entryDetailsBytes, entityKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to encrypt entry details")
	}

	// Decrypt diary key
	// Encrypt entity key with diary key
	keyNonce, encryptedEntityKey, err := encryptWithSymmetricKey(entityKey, diaryKey)
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
			return nil, errors.New("current entry change is required")
		}
		change.ObjectId = entryID
		change.ParentChangeIds = []string{current.GetChange().GetChange().GetId()}
	}

	req := &v1.PutEntryRequest{
		DiaryId: diaryID,
		EntryId: entryID,
		TopicId: protoString(params.TopicID),
		EncryptedChange: &v1.EncryptedChange{
			Change:     change,
			Encryption: encryption,
		},
	}

	response := &v1.PutEntryResponse{}
	if err := c.rpc(ctx, "io.thingsdiary.v1.entry.put", req, response); err != nil {
		return nil, err
	}
	if err := mapEntryErrorCode(response.GetErrorCode()); err != nil {
		return nil, err
	}

	return entryFromProto(response.GetEntry(), diaryKey)
}
