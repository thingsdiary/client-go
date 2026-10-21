package client

import (
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	"github.com/samber/mo"

	v1 "github.com/thingsdiary/client-go/proto"
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
	Version           uint64
}

// TopicDetails represents the plaintext content structure for topic details
type TopicDetails struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Color       string `json:"color"`
}

func topicFromProto(value *v1.Topic, diaryKey []byte) (*Topic, error) {
	if value == nil {
		return nil, errors.New("topic is required")
	}

	decryptedDetails, err := decryptProtoDetails(
		value.Encryption,
		value.Details,
		diaryKey,
		"topic details",
	)
	if err != nil {
		return nil, err
	}

	var details TopicDetails
	if err := json.Unmarshal(decryptedDetails, &details); err != nil {
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
		Version:     value.Version,
	}
	if value.DefaultTemplateId != nil {
		topic.DefaultTemplateID = mo.Some(*value.DefaultTemplateId)
	}
	if value.DeletedAt != nil {
		topic.DeletedAt = mo.Some(protoTime(value.DeletedAt))
	}

	return topic, nil
}
