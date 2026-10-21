package client

import "time"

type ChangeOperation string

const (
	ChangeOperationCreated     ChangeOperation = "created"
	ChangeOperationUpdated     ChangeOperation = "updated"
	ChangeOperationSoftDeleted ChangeOperation = "soft_deleted"
	ChangeOperationRemoved     ChangeOperation = "removed"
)

type ChangeEntityType string

const (
	ChangeEntityDiary    ChangeEntityType = "diary"
	ChangeEntityEntry    ChangeEntityType = "entry"
	ChangeEntityTopic    ChangeEntityType = "topic"
	ChangeEntityTemplate ChangeEntityType = "template"
)

type DiaryChange struct {
	ID         string
	CreatedAt  time.Time
	Operation  ChangeOperation
	EntityType ChangeEntityType
	EntityID   string
}

type GetDiaryChangesResult struct {
	Changes []*DiaryChange
	// Empty string when no more pages available
	NextPageToken string
}
