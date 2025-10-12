package client

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/samber/mo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *ClientSuite) TestEntry_GetByID() {
	t := s.T()
	ctx := context.Background()

	var login = fmt.Sprintf("test-get-entry-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	// Create diary first
	createdDiary, err := s.client.CreateDiary(ctx, CreateDiaryParams{
		Title:       "Test Diary for Entry",
		Description: "Test diary for entry retrieval",
	})
	require.NoError(t, err)
	require.NotNil(t, createdDiary)

	// Create entry
	entryID := uuid.NewString()
	createdEntry, err := s.client.PutEntry(ctx, createdDiary.ID, entryID, PutEntryParams{
		Content:       "Test entry content for retrieval",
		Archived:      false,
		Bookmarked:    true,
		PreviewHidden: false,
	})
	require.NoError(t, err)
	require.NotNil(t, createdEntry)

	// Get entry by ID
	fetchedEntry, err := s.client.GetEntryByID(ctx, createdDiary.ID, entryID)
	require.NoError(t, err)
	require.NotNil(t, fetchedEntry)

	// Verify entry content
	assert.Equal(t, entryID, fetchedEntry.ID)
	assert.Equal(t, createdDiary.ID, fetchedEntry.DiaryID)
	assert.Equal(t, "Test entry content for retrieval", fetchedEntry.Content)
	assert.False(t, fetchedEntry.TopicID.IsPresent())
	assert.False(t, fetchedEntry.Archived)
	assert.True(t, fetchedEntry.Bookmarked)
	assert.False(t, fetchedEntry.PreviewHidden)
	assert.False(t, fetchedEntry.DeletedAt.IsPresent())
	assert.Equal(t, createdEntry.Version, fetchedEntry.Version)
	assert.WithinDuration(t, createdEntry.CreatedAt, fetchedEntry.CreatedAt, time.Second)
	assert.WithinDuration(t, createdEntry.UpdatedAt, fetchedEntry.UpdatedAt, time.Second)
}

func (s *ClientSuite) TestEntry_GetByID_NotFound() {
	t := s.T()
	ctx := context.Background()

	var login = fmt.Sprintf("test-get-entry-not-found-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	// Create diary first
	createdDiary, err := s.client.CreateDiary(ctx, CreateDiaryParams{
		Title:       "Test Diary for Entry Not Found",
		Description: "Test diary for entry not found test",
	})
	require.NoError(t, err)
	require.NotNil(t, createdDiary)

	// Try to get non-existent entry
	entry, err := s.client.GetEntryByID(ctx, createdDiary.ID, uuid.NewString())
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrEntryNotFound)
	assert.Nil(t, entry)
}

func (s *ClientSuite) TestEntry_GetByID_DiaryNotFound() {
	t := s.T()
	ctx := context.Background()

	var login = fmt.Sprintf("test-get-entry-diary-not-found-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	// Try to get entry from non-existent diary
	entry, err := s.client.GetEntryByID(ctx, uuid.NewString(), uuid.NewString())
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrDiaryNotFound)
	assert.Nil(t, entry)
}

func (s *ClientSuite) TestEntry_GetByID_Unauthorized() {
	t := s.T()
	ctx := context.Background()

	// Try to get entry without authentication
	entry, err := s.client.GetEntryByID(ctx, uuid.NewString(), uuid.NewString())
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUnauthorized)
	assert.Nil(t, entry)
}

func (s *ClientSuite) TestEntry_GetByID_WithTopic() {
	t := s.T()
	ctx := context.Background()

	var login = fmt.Sprintf("test-get-entry-with-topic-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	// Create diary
	createdDiary, err := s.client.CreateDiary(ctx, CreateDiaryParams{
		Title:       "Test Diary for Entry with Topic",
		Description: "Test diary for entry with topic test",
	})
	require.NoError(t, err)
	require.NotNil(t, createdDiary)

	createdTopic, err := s.client.PutTopic(ctx, createdDiary.ID, uuid.NewString(), PutTopicParams{
		Title: "Test topic for entry with topic",
	})
	require.NoError(t, err)
	require.NotNil(t, createdTopic)

	// Create entry with topic
	entryID := uuid.NewString()
	createdEntry, err := s.client.PutEntry(ctx, createdDiary.ID, entryID, PutEntryParams{
		Content:       "Test entry content with topic",
		TopicID:       mo.Some(createdTopic.ID),
		Archived:      true,
		Bookmarked:    false,
		PreviewHidden: true,
	})
	require.NoError(t, err)
	require.NotNil(t, createdEntry)

	// Get entry by ID
	fetchedEntry, err := s.client.GetEntryByID(ctx, createdDiary.ID, entryID)
	require.NoError(t, err)
	require.NotNil(t, fetchedEntry)

	// Verify entry content including topic
	assert.Equal(t, entryID, fetchedEntry.ID)
	assert.Equal(t, createdDiary.ID, fetchedEntry.DiaryID)
	assert.Equal(t, "Test entry content with topic", fetchedEntry.Content)
	assert.True(t, fetchedEntry.TopicID.IsPresent())
	assert.Equal(t, createdTopic.ID, fetchedEntry.TopicID.MustGet())
	assert.True(t, fetchedEntry.Archived)
	assert.False(t, fetchedEntry.Bookmarked)
	assert.True(t, fetchedEntry.PreviewHidden)
	assert.False(t, fetchedEntry.DeletedAt.IsPresent())
	assert.Equal(t, createdEntry.Version, fetchedEntry.Version)
}
