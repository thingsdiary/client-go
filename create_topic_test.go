package client

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *ClientSuite) TestTopic_CreateTopic() {
	t := s.T()
	ctx := context.Background()

	// Arrange
	var login = fmt.Sprintf("test-login-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	createdDiary, err := s.client.CreateDiary(ctx, CreateDiaryParams{
		Title:       "Test Diary for Topic",
		Description: "Test diary for topic creation",
	})
	require.NoError(t, err)
	require.NotNil(t, createdDiary)

	// Act
	createdTopic, err := s.client.CreateTopic(ctx, createdDiary.ID, CreateTopicParams{
		Title:       "My Personal Topic",
		Description: "A topic for personal thoughts",
		Color:       "#FF5733",
	})

	// Assert
	require.NoError(t, err)
	require.NotNil(t, createdTopic)

	assert.NotEmpty(t, createdTopic.ID, "Topic ID should be generated")
	assert.Equal(t, createdDiary.ID, createdTopic.DiaryID)
	assert.Equal(t, "My Personal Topic", createdTopic.Title)
	assert.Equal(t, "A topic for personal thoughts", createdTopic.Description)
	assert.Equal(t, "#FF5733", createdTopic.Color)
	assert.Greater(t, createdTopic.Version, uint64(0))
	assert.WithinDuration(t, time.Now(), createdTopic.CreatedAt, 5*time.Second)
	assert.WithinDuration(t, time.Now(), createdTopic.UpdatedAt, 5*time.Second)
}

func (s *ClientSuite) TestTopic_CreateTopic_EmptyFields() {
	t := s.T()
	ctx := context.Background()

	// Arrange
	var login = fmt.Sprintf("test-login-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	createdDiary, err := s.client.CreateDiary(ctx, CreateDiaryParams{
		Title:       "Test Diary for Empty Topic",
		Description: "Test diary for topic with empty fields",
	})
	require.NoError(t, err)

	// Act
	createdTopic, err := s.client.CreateTopic(ctx, createdDiary.ID, CreateTopicParams{
		Title:       "",
		Description: "",
		Color:       "",
	})

	// Assert
	require.NoError(t, err)
	require.NotNil(t, createdTopic)

	assert.NotEmpty(t, createdTopic.ID)
	assert.Equal(t, createdDiary.ID, createdTopic.DiaryID)
	assert.Equal(t, "", createdTopic.Title)
	assert.Equal(t, "", createdTopic.Description)
	assert.Equal(t, "", createdTopic.Color)
}

func (s *ClientSuite) TestTopic_CreateTopic_DiaryNotFound() {
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
	topic, err := s.client.CreateTopic(ctx, nonExistentDiaryID, CreateTopicParams{
		Title:       "Test Topic",
		Description: "Test description",
		Color:       "#FF0000",
	})

	// Assert
	require.Error(t, err)
	require.Nil(t, topic)
	assert.ErrorIs(t, err, ErrDiaryNotFound)
}

func (s *ClientSuite) TestTopic_CreateTopic_Unauthorized() {
	t := s.T()
	ctx := context.Background()

	// Arrange
	// No authentication

	// Act
	topic, err := s.client.CreateTopic(ctx, "some-diary-id", CreateTopicParams{
		Title:       "Test Topic",
		Description: "Test description",
		Color:       "#FF0000",
	})

	// Assert
	require.Error(t, err)
	require.Nil(t, topic)
	assert.ErrorIs(t, err, ErrUnauthorized)
}
