package client

import (
	"context"
	"fmt"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *ClientSuite) TestDiary_GetDiaries() {
	t := s.T()
	ctx := context.Background()

	// Arrange
	var login = fmt.Sprintf("test-login-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	// Create first diary
	createdDiary1, err := s.client.CreateDiary(ctx, CreateDiaryParams{
		Title:       "My First Test Diary",
		Description: "First test diary for GetDiaries test",
	})
	require.NoError(t, err)
	require.NotNil(t, createdDiary1)

	// Create second diary
	createdDiary2, err := s.client.CreateDiary(ctx, CreateDiaryParams{
		Title:       "My Second Test Diary",
		Description: "Second test diary for GetDiaries test",
	})
	require.NoError(t, err)
	require.NotNil(t, createdDiary2)

	// Act
	diaries, err := s.client.GetDiaries(ctx)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, diaries)
	require.Len(t, diaries, 2)

	// Find diaries by ID
	var foundDiary1, foundDiary2 *Diary
	for _, diary := range diaries {
		switch diary.ID {
		case createdDiary1.ID:
			foundDiary1 = diary
		case createdDiary2.ID:
			foundDiary2 = diary
		}
	}

	require.NotNil(t, foundDiary1, "First diary should be found in the list")
	require.NotNil(t, foundDiary2, "Second diary should be found in the list")

	// Verify first diary
	assert.Equal(t, createdDiary1.ID, foundDiary1.ID)
	assert.Equal(t, "My First Test Diary", foundDiary1.Title)
	assert.Equal(t, "First test diary for GetDiaries test", foundDiary1.Description)
	assert.Equal(t, createdDiary1.Version, foundDiary1.Version)
	assert.WithinDuration(t, createdDiary1.CreatedAt, foundDiary1.CreatedAt, 1*time.Second)
	assert.WithinDuration(t, createdDiary1.UpdatedAt, foundDiary1.UpdatedAt, 1*time.Second)

	// Verify second diary
	assert.Equal(t, createdDiary2.ID, foundDiary2.ID)
	assert.Equal(t, "My Second Test Diary", foundDiary2.Title)
	assert.Equal(t, "Second test diary for GetDiaries test", foundDiary2.Description)
	assert.Equal(t, createdDiary2.Version, foundDiary2.Version)
	assert.WithinDuration(t, createdDiary2.CreatedAt, foundDiary2.CreatedAt, 1*time.Second)
	assert.WithinDuration(t, createdDiary2.UpdatedAt, foundDiary2.UpdatedAt, 1*time.Second)
}

func (s *ClientSuite) TestDiary_GetDiaries_EmptyList() {
	t := s.T()
	ctx := context.Background()

	// Arrange
	var login = fmt.Sprintf("test-login-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)
	// (no diaries created)

	// Act
	diaries, err := s.client.GetDiaries(ctx)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, diaries)
	require.Empty(t, diaries, "Should return empty list when account has no diaries")
}

func (s *ClientSuite) TestDiary_GetDiaries_Unauthorized() {
	t := s.T()
	ctx := context.Background()

	// Arrange
	// (no authentication setup)

	// Act
	diaries, err := s.client.GetDiaries(ctx)

	// Assert
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUnauthorized)
	assert.Nil(t, diaries)
}
