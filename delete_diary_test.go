package client

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *ClientSuite) TestDiary_DeleteDiary() {
	t := s.T()
	ctx := context.Background()

	// Arrange: Register and authenticate user
	var login = fmt.Sprintf("test-delete-diary-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	// Create diary to delete
	createdDiary, err := s.client.CreateDiary(ctx, CreateDiaryParams{
		Title:       "Diary to Delete",
		Description: "This diary will be deleted",
	})
	require.NoError(t, err)
	require.NotNil(t, createdDiary)

	// Act: Delete diary
	err = s.client.DeleteDiary(ctx, createdDiary.ID)

	// Assert: Deletion was successful
	require.NoError(t, err)

	// Verify diary no longer exists
	_, err = s.client.GetDiaryByID(ctx, createdDiary.ID)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrDiaryNotFound)
}

func (s *ClientSuite) TestDiary_DeleteDiary_NotFound() {
	t := s.T()
	ctx := context.Background()

	// Arrange: Register and authenticate user
	var login = fmt.Sprintf("test-delete-diary-not-found-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	// Act: Try to delete non-existent diary
	err = s.client.DeleteDiary(ctx, uuid.NewString())

	// Assert: Returns diary not found error
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrDiaryNotFound)
}

func (s *ClientSuite) TestDiary_DeleteDiary_Unauthorized() {
	t := s.T()
	ctx := context.Background()

	// Act: Try to delete diary without authentication
	err := s.client.DeleteDiary(ctx, uuid.NewString())

	// Assert: Returns unauthorized error
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUnauthorized)
}

func (s *ClientSuite) TestDiary_DeleteDiary_MultipleOperations() {
	t := s.T()
	ctx := context.Background()

	// Arrange: Register and authenticate user
	var login = fmt.Sprintf("test-delete-multiple-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	// Create multiple diaries
	diary1, err := s.client.CreateDiary(ctx, CreateDiaryParams{
		Title:       "First Diary",
		Description: "First diary to delete",
	})
	require.NoError(t, err)

	diary2, err := s.client.CreateDiary(ctx, CreateDiaryParams{
		Title:       "Second Diary",
		Description: "Second diary to delete",
	})
	require.NoError(t, err)

	// Act: Delete first diary
	err = s.client.DeleteDiary(ctx, diary1.ID)
	require.NoError(t, err)

	// Assert: First diary is deleted, second still exists
	_, err = s.client.GetDiaryByID(ctx, diary1.ID)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrDiaryNotFound)

	fetchedDiary2, err := s.client.GetDiaryByID(ctx, diary2.ID)
	require.NoError(t, err)
	assert.Equal(t, diary2.ID, fetchedDiary2.ID)
	assert.Equal(t, "Second Diary", fetchedDiary2.Title)
}

func (s *ClientSuite) TestDiary_DeleteDiary_ThenUpdateDeleted() {
	t := s.T()
	ctx := context.Background()

	// Arrange: Register and authenticate user
	var login = fmt.Sprintf("test-delete-then-update-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	// Create diary
	createdDiary, err := s.client.CreateDiary(ctx, CreateDiaryParams{
		Title:       "Diary to Delete and Update",
		Description: "This diary will be deleted then updated",
	})
	require.NoError(t, err)
	require.NotNil(t, createdDiary)

	// Delete diary
	err = s.client.DeleteDiary(ctx, createdDiary.ID)
	require.NoError(t, err)

	// Act: Try to update deleted diary
	_, err = s.client.PutDiary(ctx, createdDiary.ID, PutDiaryParams{
		Title:       "Updated Title",
		Description: "Updated Description",
	})

	// Assert: Should return diary not found error
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrDiaryNotFound)
}
