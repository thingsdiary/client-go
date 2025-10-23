package client

import (
	"context"
	"fmt"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *ClientSuite) TestGetActiveDiaryKey_Success() {
	t := s.T()
	ctx := context.Background()

	// Register and authenticate
	var login = fmt.Sprintf("test-login-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	// Create a diary
	diary, err := s.client.CreateDiary(ctx, CreateDiaryParams{
		Title:       "Test Diary for Keys",
		Description: "Testing key retrieval",
	})
	require.NoError(t, err)
	require.NotNil(t, diary)

	// Get active key
	key, err := s.client.getActiveDiaryKey(ctx, diary.ID)
	require.NoError(t, err)
	require.NotNil(t, key)

	// Verify key structure
	assert.NotEmpty(t, key.Id)
	assert.NotEmpty(t, key.Value)
	assert.Equal(t, "active", string(key.Status))
}

func (s *ClientSuite) TestGetActiveDiaryKey_NotFound() {
	t := s.T()
	ctx := context.Background()

	// Register and authenticate
	var login = fmt.Sprintf("test-login-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	// Try to get keys for non-existent diary (with invalid UUID)
	key, err := s.client.getActiveDiaryKey(ctx, "non-existent-diary-id")
	require.Error(t, err)
	assert.Nil(t, key)
	// Server returns 500 for invalid UUID format, which is expected
}

func (s *ClientSuite) TestGetActiveDiaryKey_AfterDeletion() {
	t := s.T()
	ctx := context.Background()

	// Register and authenticate
	var login = fmt.Sprintf("test-login-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	// Create a diary
	diary, err := s.client.CreateDiary(ctx, CreateDiaryParams{
		Title:       "Diary to Delete",
		Description: "This diary will be deleted",
	})
	require.NoError(t, err)
	require.NotNil(t, diary)

	// Verify we can get keys before deletion
	key, err := s.client.getActiveDiaryKey(ctx, diary.ID)
	require.NoError(t, err)
	require.NotNil(t, key)

	// Delete the diary
	err = s.client.DeleteDiary(ctx, diary.ID)
	require.NoError(t, err)

	// Try to get keys for deleted diary - should return 404
	key, err = s.client.getActiveDiaryKey(ctx, diary.ID)
	require.Error(t, err)
	assert.Nil(t, key)
	assert.Equal(t, ErrDiaryNotFound, err)
}
