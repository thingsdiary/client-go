package client

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *ClientSuite) TestEntry_DeleteEntry() {
	t := s.T()
	ctx := context.Background()

	// Arrange: Register and authenticate user
	var login = fmt.Sprintf("test-delete-entry-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	// Create diary
	createdDiary, err := s.client.CreateDiary(ctx, CreateDiaryParams{
		Title:       "Test Diary",
		Description: "Diary for entry deletion test",
	})
	require.NoError(t, err)
	require.NotNil(t, createdDiary)

	createdEntry, err := s.client.CreateEntry(ctx, createdDiary.ID, CreateEntryParams{
		Content: "Entry to be deleted",
	})
	require.NoError(t, err)
	require.NotNil(t, createdEntry)

	// Act: Delete entry
	err = s.client.DeleteEntry(ctx, createdDiary.ID, createdEntry.ID)
	require.NoError(t, err)

	// Assert
	gotEntry, err := s.client.GetEntryByID(ctx, createdDiary.ID, createdEntry.ID)
	require.NoError(t, err)
	assert.True(t, gotEntry.DeletedAt.IsPresent())
}

func (s *ClientSuite) TestEntry_DeleteEntry_NotFound() {
	t := s.T()
	ctx := context.Background()

	// Arrange: Register and authenticate user
	var login = fmt.Sprintf("test-delete-entry-not-found-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	// Create diary
	createdDiary, err := s.client.CreateDiary(ctx, CreateDiaryParams{
		Title:       "Test Diary",
		Description: "Diary for entry not found test",
	})
	require.NoError(t, err)
	require.NotNil(t, createdDiary)

	// Act: Try to delete non-existent entry
	err = s.client.DeleteEntry(ctx, createdDiary.ID, uuid.NewString())

	// Assert: Returns entry not found error
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrEntryNotFound)
}

func (s *ClientSuite) TestEntry_DeleteEntry_Unauthorized() {
	t := s.T()
	ctx := context.Background()

	// Act: Try to delete entry without authentication
	err := s.client.DeleteEntry(ctx, uuid.NewString(), uuid.NewString())

	// Assert: Returns unauthorized error
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUnauthorized)
}
