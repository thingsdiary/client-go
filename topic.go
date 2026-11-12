package client

import (
	"time"

	"github.com/pkg/errors"
	"github.com/samber/mo"
	v1 "github.com/thingsdiary/client-go/proto"
	"google.golang.org/protobuf/proto"
)

// Topic represents a plaintext topic that users work with
type Topic struct {
	ID                string
	DiaryID           string
	Title             string
	Description       string
	Color             string
	DefaultTemplateID mo.Option[string]
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         mo.Option[time.Time]
	RemovedAt         mo.Option[time.Time]
	ChangeID          string
}

// TopicDetails represents the plaintext content structure for topic details
type TopicDetails struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Color       string `json:"color"`
}

func topicFromProto(value *v1.EncryptedTopic, diaryKey []byte) (*Topic, error) {
	if value == nil {
		return nil, errors.New("topic is required")
	}

	change := value.GetChange()
	if change == nil || change.GetChange() == nil || change.GetEncryption() == nil {
		return nil, errors.New("topic change is required")
	}

	nonce, encryptedDetails, err := UnmarshalChangePayload(change.GetChange().GetPayload())
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode topic change payload")
	}
	entityKey, err := decryptWithSymmetricKey(
		change.GetEncryption().GetEncryptedKeyNonce(),
		change.GetEncryption().GetEncryptedKeyData(),
		diaryKey,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decrypt topic entity key")
	}
	decryptedDetails, err := decryptWithSymmetricKey(nonce, encryptedDetails, entityKey)
	if err != nil {
		return nil, err
	}

	var details v1.TopicDetails
	if err := proto.Unmarshal(decryptedDetails, &details); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal topic details")
	}

	topic := &Topic{
		ID:          value.Id,
		DiaryID:     value.DiaryId,
		Title:       details.Title,
		Description: details.Description,
		Color:       details.Color,
		CreatedAt:   protoTime(value.CreatedAt),
		UpdatedAt:   protoTime(value.UpdatedAt),
		ChangeID:    change.GetChange().GetId(),
	}
	if value.DefaultTemplateId != nil {
		topic.DefaultTemplateID = mo.Some(*value.DefaultTemplateId)
	}
	if value.DeletedAt != nil {
		topic.DeletedAt = mo.Some(protoTime(value.DeletedAt))
	}
	if value.RemovedAt != nil {
		topic.RemovedAt = mo.Some(protoTime(value.RemovedAt))
	}

	return topic, nil
}
