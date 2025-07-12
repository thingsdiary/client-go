package client

import (
	"time"

	"github.com/samber/mo"
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
