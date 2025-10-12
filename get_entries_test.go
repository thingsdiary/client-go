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

func (s *ClientSuite) TestEntry_GetEntries() {
	t := s.T()
	ctx := context.Background()

	// Arrange
	var login = fmt.Sprintf("test-login-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	// Create diary
	diary, err := s.client.CreateDiary(ctx, CreateDiaryParams{
		Title:       "Test Diary",
		Description: "Diary for entries testing",
	})
	require.NoError(t, err)
	require.NotNil(t, diary)

	// Create first entry
	entryID1 := uuid.NewString()
	createdEntry1, err := s.client.PutEntry(ctx, diary.ID, entryID1, PutEntryParams{
		Content:       "My first entry content",
		TopicID:       mo.None[string](),
		Archived:      false,
		Bookmarked:    true,
		PreviewHidden: false,
	})
	require.NoError(t, err)
	require.NotNil(t, createdEntry1)

	// Create second entry
	entryID2 := uuid.NewString()
	createdEntry2, err := s.client.PutEntry(ctx, diary.ID, entryID2, PutEntryParams{
		Content:       "My second entry content",
		TopicID:       mo.None[string](),
		Archived:      true,
		Bookmarked:    false,
		PreviewHidden: true,
	})
	require.NoError(t, err)
	require.NotNil(t, createdEntry2)

	// Act
	entries, err := s.client.GetEntries(ctx, diary.ID)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, entries)
	require.Len(t, entries, 2)

	// Find entries by ID
	var foundEntry1, foundEntry2 *Entry
	for _, entry := range entries {
		switch entry.ID {
		case entryID1:
			foundEntry1 = entry
		case entryID2:
			foundEntry2 = entry
		}
	}

	require.NotNil(t, foundEntry1, "First entry should be found in the list")
	require.NotNil(t, foundEntry2, "Second entry should be found in the list")

	// Verify first entry
	assert.Equal(t, entryID1, foundEntry1.ID)
	assert.Equal(t, diary.ID, foundEntry1.DiaryID)
	assert.Equal(t, "My first entry content", foundEntry1.Content)
	assert.False(t, foundEntry1.TopicID.IsPresent())
	assert.False(t, foundEntry1.Archived)
	assert.True(t, foundEntry1.Bookmarked)
	assert.False(t, foundEntry1.PreviewHidden)
	assert.Equal(t, createdEntry1.Version, foundEntry1.Version)
	assert.WithinDuration(t, createdEntry1.CreatedAt, foundEntry1.CreatedAt, 1*time.Second)
	assert.WithinDuration(t, createdEntry1.UpdatedAt, foundEntry1.UpdatedAt, 1*time.Second)

	// Verify second entry
	assert.Equal(t, entryID2, foundEntry2.ID)
	assert.Equal(t, diary.ID, foundEntry2.DiaryID)
	assert.Equal(t, "My second entry content", foundEntry2.Content)
	assert.False(t, foundEntry2.TopicID.IsPresent())
	assert.True(t, foundEntry2.Archived)
	assert.False(t, foundEntry2.Bookmarked)
	assert.True(t, foundEntry2.PreviewHidden)
	assert.Equal(t, createdEntry2.Version, foundEntry2.Version)
	assert.WithinDuration(t, createdEntry2.CreatedAt, foundEntry2.CreatedAt, 1*time.Second)
	assert.WithinDuration(t, createdEntry2.UpdatedAt, foundEntry2.UpdatedAt, 1*time.Second)
}

func (s *ClientSuite) TestEntry_GetEntries_EmptyList() {
	t := s.T()
	ctx := context.Background()

	// Arrange
	var login = fmt.Sprintf("test-login-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	// Create diary (without entries)
	diary, err := s.client.CreateDiary(ctx, CreateDiaryParams{
		Title:       "Test Diary",
		Description: "Diary for entries testing",
	})
	require.NoError(t, err)

	// Act
	entries, err := s.client.GetEntries(ctx, diary.ID)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, entries)
	require.Empty(t, entries, "Should return empty list when diary has no entries")
}

func (s *ClientSuite) TestEntry_GetEntries_Unauthorized() {
	t := s.T()
	ctx := context.Background()

	// Arrange
	// (no authentication setup)

	// Act
	entries, err := s.client.GetEntries(ctx, "some-diary-id")

	// Assert
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUnauthorized)
	assert.Nil(t, entries)
}

func (s *ClientSuite) TestEntry_GetEntries_DiaryNotFound() {
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
	entries, err := s.client.GetEntries(ctx, nonExistentDiaryID)

	// Assert
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrDiaryNotFound)
	assert.Nil(t, entries)
}
