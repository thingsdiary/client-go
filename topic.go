package client

import (
	"time"

	"github.com/samber/mo"
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
