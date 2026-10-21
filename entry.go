package client

import (
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	"github.com/samber/mo"
	v1 "github.com/thingsdiary/client-go/proto"
)

// Entry represents a plaintext entry that users work with
type Entry struct {
	ID            string
	DiaryID       string
	Content       string
	TopicID       mo.Option[string]
	Archived      bool
	Bookmarked    bool
	PreviewHidden bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     mo.Option[time.Time]
	Version       uint64
}

// EntryDetails represents the plaintext content structure for entry details
type EntryDetails struct {
	Content       string `json:"content"`
	Archived      bool   `json:"archived"`
	Bookmarked    bool   `json:"bookmarked"`
	PreviewHidden bool   `json:"preview_hidden"`
}

func entryFromProto(value *v1.Entry, diaryKey []byte) (*Entry, error) {
	if value == nil {
		return nil, errors.New("entry is required")
	}

	decryptedDetails, err := decryptProtoDetails(
		value.Encryption,
		value.Details,
		diaryKey,
		"entry details",
	)
	if err != nil {
		return nil, err
	}

	var details EntryDetails
	if err := json.Unmarshal(decryptedDetails, &details); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal entry details")
	}

	entry := &Entry{
		ID:            value.Id,
		DiaryID:       value.DiaryId,
		Content:       details.Content,
		Archived:      details.Archived,
		Bookmarked:    details.Bookmarked,
		PreviewHidden: details.PreviewHidden,
		CreatedAt:     protoTime(value.CreatedAt),
		UpdatedAt:     protoTime(value.UpdatedAt),
		Version:       value.Version,
	}
	if value.TopicId != nil {
		entry.TopicID = mo.Some(*value.TopicId)
	}

	if value.DeletedAt != nil {
		entry.DeletedAt = mo.Some(protoTime(value.DeletedAt))
	}

	return entry, nil
}
