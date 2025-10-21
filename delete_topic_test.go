package client

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *ClientSuite) TestTopic_DeleteTopic() {
	t := s.T()
	ctx := context.Background()

	// Arrange: Register and authenticate user
	var login = fmt.Sprintf("test-delete-topic-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	// Create diary
	createdDiary, err := s.client.CreateDiary(ctx, CreateDiaryParams{
		Title:       "Test Diary",
		Description: "Diary for topic deletion test",
	})
	require.NoError(t, err)
	require.NotNil(t, createdDiary)

	// Create topic to delete
	createdTopic, err := s.client.CreateTopic(ctx, createdDiary.ID, CreateTopicParams{
		Title:       "Topic to be deleted",
		Description: "This topic will be deleted",
		Color:       "#FF0000",
	})
	require.NoError(t, err)
	require.NotNil(t, createdTopic)

	// Act: Delete topic
	err = s.client.DeleteTopic(ctx, createdDiary.ID, createdTopic.ID)
	require.NoError(t, err)

	// Assert

	gotTopic, err := s.client.GetTopicByID(ctx, createdDiary.ID, createdTopic.ID)
	require.NoError(t, err)
	assert.True(t, gotTopic.DeletedAt.IsPresent())
}

func (s *ClientSuite) TestTopic_DeleteTopic_NotFound() {
	t := s.T()
	ctx := context.Background()

	// Arrange: Register and authenticate user
	var login = fmt.Sprintf("test-delete-topic-not-found-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	// Create diary
	createdDiary, err := s.client.CreateDiary(ctx, CreateDiaryParams{
		Title:       "Test Diary",
		Description: "Diary for topic not found test",
	})
	require.NoError(t, err)
	require.NotNil(t, createdDiary)

	// Act: Try to delete non-existent topic
	err = s.client.DeleteTopic(ctx, createdDiary.ID, uuid.NewString())

	// Assert: Returns topic not found error
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrTopicNotFound)
}

func (s *ClientSuite) TestTopic_DeleteTopic_Unauthorized() {
	t := s.T()
	ctx := context.Background()

	// Act: Try to delete topic without authentication
	err := s.client.DeleteTopic(ctx, uuid.NewString(), uuid.NewString())

	// Assert: Returns unauthorized error
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUnauthorized)
}
