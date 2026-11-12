package client

import (
	"time"

	"github.com/pkg/errors"
	"github.com/samber/mo"
	v1 "github.com/thingsdiary/client-go/proto"
	"google.golang.org/protobuf/proto"
)

// Template represents a plaintext template that users work with
type Template struct {
	ID        string
	DiaryID   string
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt mo.Option[time.Time]
	RemovedAt mo.Option[time.Time]
	Version   uint64
}

// TemplateDetails represents the plaintext content structure for template details
type TemplateDetails struct {
	Content string `json:"content"`
}

func templateFromProto(value *v1.EncryptedTemplate, diaryKey []byte) (*Template, error) {
	if value == nil {
		return nil, errors.New("template is required")
	}

	decryptedDetails, err := decryptProtoDetails(
		value.Encryption,
		value.Details,
		diaryKey,
		"template details",
	)
	if err != nil {
		return nil, err
	}

	var details v1.TemplateDetails
	if err := proto.Unmarshal(decryptedDetails, &details); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal template details")
	}

	template := &Template{
		ID:        value.Id,
		DiaryID:   value.DiaryId,
		Content:   details.Content,
		CreatedAt: protoTime(value.CreatedAt),
		UpdatedAt: protoTime(value.UpdatedAt),
		Version:   value.Version,
	}
	if value.DeletedAt != nil {
		template.DeletedAt = mo.Some(protoTime(value.DeletedAt))
	}
	if value.RemovedAt != nil {
		template.RemovedAt = mo.Some(protoTime(value.RemovedAt))
	}

	return template, nil
}
