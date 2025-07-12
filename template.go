package client

import (
	"time"

	"github.com/samber/mo"
)

// Template represents a plaintext template that users work with
type Template struct {
	ID        string
	DiaryID   string
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt mo.Option[time.Time]
	Version   uint64
}

// TemplateDetails represents the plaintext content structure for template details
type TemplateDetails struct {
	Content string `json:"content"`
}
