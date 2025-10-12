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

func (s *ClientSuite) TestEntry_CreateEntry() {
	t := s.T()
	ctx := context.Background()

	// Arrange
	var login = fmt.Sprintf("test-login-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	createdDiary, err := s.client.CreateDiary(ctx, CreateDiaryParams{
		Title:       "Test Diary for Entry",
		Description: "Test diary for entry creation",
	})
	require.NoError(t, err)
	require.NotNil(t, createdDiary)

	// Act
	createdEntry, err := s.client.CreateEntry(ctx, createdDiary.ID, CreateEntryParams{
		Content:       "This is a test entry content",
		TopicID:       mo.None[string](),
		Archived:      false,
		Bookmarked:    true,
		PreviewHidden: false,
	})

	// Assert
	require.NoError(t, err)
	require.NotNil(t, createdEntry)

	assert.NotEmpty(t, createdEntry.ID, "Entry ID should be generated")
	assert.Equal(t, createdDiary.ID, createdEntry.DiaryID)
	assert.Equal(t, "This is a test entry content", createdEntry.Content)
	assert.True(t, createdEntry.TopicID.IsAbsent())
	assert.False(t, createdEntry.Archived)
	assert.True(t, createdEntry.Bookmarked)
	assert.False(t, createdEntry.PreviewHidden)
	assert.Greater(t, createdEntry.Version, uint64(0))
	assert.WithinDuration(t, time.Now(), createdEntry.CreatedAt, 5*time.Second)
	assert.WithinDuration(t, time.Now(), createdEntry.UpdatedAt, 5*time.Second)
	assert.True(t, createdEntry.DeletedAt.IsAbsent())
}

func (s *ClientSuite) TestEntry_CreateEntry_WithTopic() {
	t := s.T()
	ctx := context.Background()

	// Arrange
	var login = fmt.Sprintf("test-login-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	createdDiary, err := s.client.CreateDiary(ctx, CreateDiaryParams{
		Title:       "Test Diary for Entry with Topic",
		Description: "Test diary for entry creation with topic",
	})
	require.NoError(t, err)

	createdTopic, err := s.client.CreateTopic(ctx, createdDiary.ID, CreateTopicParams{
		Title:       "Test Topic",
		Description: "Test topic for entry",
		Color:       "#FF0000",
	})
	require.NoError(t, err)

	// Act
	createdEntry, err := s.client.CreateEntry(ctx, createdDiary.ID, CreateEntryParams{
		Content:       "Entry with topic",
		TopicID:       mo.Some(createdTopic.ID),
		Archived:      true,
		Bookmarked:    false,
		PreviewHidden: true,
	})

	// Assert
	require.NoError(t, err)
	require.NotNil(t, createdEntry)

	assert.NotEmpty(t, createdEntry.ID)
	assert.Equal(t, createdDiary.ID, createdEntry.DiaryID)
	assert.Equal(t, "Entry with topic", createdEntry.Content)
	assert.True(t, createdEntry.TopicID.IsPresent())
	assert.Equal(t, createdTopic.ID, createdEntry.TopicID.MustGet())
	assert.True(t, createdEntry.Archived)
	assert.False(t, createdEntry.Bookmarked)
	assert.True(t, createdEntry.PreviewHidden)
}

func (s *ClientSuite) TestEntry_CreateEntry_DiaryNotFound() {
	t := s.T()
	ctx := context.Background()

	// Arrange
	var login = fmt.Sprintf("test-login-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	nonExistentDiaryID := uuid.NewString()

	// Act
	entry, err := s.client.CreateEntry(ctx, nonExistentDiaryID, CreateEntryParams{
		Content: "Test content",
	})

	// Assert
	require.Error(t, err)
	require.Nil(t, entry)
	assert.ErrorIs(t, err, ErrDiaryNotFound)
}

func (s *ClientSuite) TestEntry_CreateEntry_Unauthorized() {
	t := s.T()
	ctx := context.Background()

	// Arrange
	// No authentication

	// Act
	entry, err := s.client.CreateEntry(ctx, "some-diary-id", CreateEntryParams{
		Content: "Test content",
	})

	// Assert
	require.Error(t, err)
	require.Nil(t, entry)
	assert.ErrorIs(t, err, ErrUnauthorized)
}
