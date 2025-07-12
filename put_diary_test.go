package client

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *ClientSuite) TestDiary_PutDiary() {
	t := s.T()
	ctx := context.Background()

	// Arrange: Register and authenticate user
	var login = fmt.Sprintf("test-put-diary-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	// Create initial diary
	initialDiary, err := s.client.CreateDiary(ctx, CreateDiaryParams{
		Title:       "Initial Title",
		Description: "Initial Description",
	})
	require.NoError(t, err)
	require.NotNil(t, initialDiary)

	// Act: Update diary
	updatedDiary, err := s.client.PutDiary(ctx, initialDiary.ID, PutDiaryParams{
		Title:       "Updated Title",
		Description: "Updated Description",
	})

	// Assert: Verify update was successful
	require.NoError(t, err)
	require.NotNil(t, updatedDiary)

	assert.Equal(t, initialDiary.ID, updatedDiary.ID)
	assert.Equal(t, "Updated Title", updatedDiary.Title)
	assert.Equal(t, "Updated Description", updatedDiary.Description)
	assert.WithinDuration(t, initialDiary.CreatedAt, updatedDiary.CreatedAt, time.Second)
	assert.WithinDuration(t, updatedDiary.UpdatedAt, initialDiary.UpdatedAt, time.Second)
	assert.Greater(t, updatedDiary.Version, initialDiary.Version)

	// Verify diary was actually updated by fetching it again
	fetchedDiary, err := s.client.GetDiaryByID(ctx, initialDiary.ID)
	require.NoError(t, err)
	require.NotNil(t, fetchedDiary)

	assert.Equal(t, "Updated Title", fetchedDiary.Title)
	assert.Equal(t, "Updated Description", fetchedDiary.Description)
	assert.Greater(t, fetchedDiary.Version, initialDiary.Version)
}

func (s *ClientSuite) TestDiary_PutDiary_NotFound() {
	t := s.T()
	ctx := context.Background()

	// Arrange: Register and authenticate user
	var login = fmt.Sprintf("test-put-diary-not-found-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	// Act: Try to update non-existent diary
	diary, err := s.client.PutDiary(ctx, uuid.NewString(), PutDiaryParams{
		Title:       "Test Title",
		Description: "Test Description",
	})

	// Assert: Should return error
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrDiaryNotFound)
	assert.Nil(t, diary)
}

func (s *ClientSuite) TestDiary_PutDiary_NotAuthenticated() {
	t := s.T()
	ctx := context.Background()

	// Act: Try to update diary without authentication
	diary, err := s.client.PutDiary(ctx, uuid.NewString(), PutDiaryParams{
		Title:       "Test Title",
		Description: "Test Description",
	})

	// Assert: Should return unauthorized error
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUnauthorized)
	assert.Nil(t, diary)
}

func (s *ClientSuite) TestDiary_PutDiary_PartialUpdate() {
	t := s.T()
	ctx := context.Background()

	// Arrange: Register and authenticate user
	var login = fmt.Sprintf("test-put-diary-partial-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	// Create initial diary
	initialDiary, err := s.client.CreateDiary(ctx, CreateDiaryParams{
		Title:       "Initial Title",
		Description: "Initial Description",
	})
	require.NoError(t, err)
	require.NotNil(t, initialDiary)

	// Act: Update only title
	updatedDiary, err := s.client.PutDiary(ctx, initialDiary.ID, PutDiaryParams{
		Title:       "New Title Only",
		Description: "New Description Too",
	})

	// Assert: Both fields should be updated
	require.NoError(t, err)
	require.NotNil(t, updatedDiary)

	assert.Equal(t, initialDiary.ID, updatedDiary.ID)
	assert.Equal(t, "New Title Only", updatedDiary.Title)
	assert.Equal(t, "New Description Too", updatedDiary.Description)
	assert.Greater(t, updatedDiary.Version, initialDiary.Version)
}

func (s *ClientSuite) TestDiary_PutDiary_EmptyFields() {
	t := s.T()
	ctx := context.Background()

	// Arrange: Register and authenticate user
	var login = fmt.Sprintf("test-put-diary-empty-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	// Create initial diary
	initialDiary, err := s.client.CreateDiary(ctx, CreateDiaryParams{
		Title:       "Initial Title",
		Description: "Initial Description",
	})
	require.NoError(t, err)
	require.NotNil(t, initialDiary)

	// Act: Update with empty fields
	updatedDiary, err := s.client.PutDiary(ctx, initialDiary.ID, PutDiaryParams{
		Title:       "",
		Description: "",
	})

	// Assert: Empty fields should be allowed and applied
	require.NoError(t, err)
	require.NotNil(t, updatedDiary)

	assert.Equal(t, initialDiary.ID, updatedDiary.ID)
	assert.Equal(t, "", updatedDiary.Title)
	assert.Equal(t, "", updatedDiary.Description)
	assert.Greater(t, updatedDiary.Version, initialDiary.Version)
}
