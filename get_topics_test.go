package client

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *ClientSuite) TestTopic_GetTopics() {
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
		Description: "Diary for topics testing",
	})
	require.NoError(t, err)
	require.NotNil(t, diary)

	// Create first topic
	topicID1 := uuid.NewString()
	createdTopic1, err := s.client.PutTopic(ctx, diary.ID, topicID1, PutTopicParams{
		Title:       "My First Topic",
		Description: "First topic description",
		Color:       "#FF5733",
	})
	require.NoError(t, err)
	require.NotNil(t, createdTopic1)

	// Create second topic
	topicID2 := uuid.NewString()
	createdTopic2, err := s.client.PutTopic(ctx, diary.ID, topicID2, PutTopicParams{
		Title:       "My Second Topic",
		Description: "Second topic description",
		Color:       "#33FF57",
	})
	require.NoError(t, err)
	require.NotNil(t, createdTopic2)

	// Act
	topics, err := s.client.GetTopics(ctx, diary.ID)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, topics)
	require.Len(t, topics, 2)

	// Find topics by ID
	var foundTopic1, foundTopic2 *Topic
	for _, topic := range topics {
		switch topic.ID {
		case topicID1:
			foundTopic1 = topic
		case topicID2:
			foundTopic2 = topic
		}
	}

	require.NotNil(t, foundTopic1, "First topic should be found in the list")
	require.NotNil(t, foundTopic2, "Second topic should be found in the list")

	// Verify first topic
	assert.Equal(t, topicID1, foundTopic1.ID)
	assert.Equal(t, diary.ID, foundTopic1.DiaryID)
	assert.Equal(t, "My First Topic", foundTopic1.Title)
	assert.Equal(t, "First topic description", foundTopic1.Description)
	assert.Equal(t, "#FF5733", foundTopic1.Color)
	assert.Equal(t, createdTopic1.Version, foundTopic1.Version)
	assert.WithinDuration(t, createdTopic1.CreatedAt, foundTopic1.CreatedAt, 1*time.Second)
	assert.WithinDuration(t, createdTopic1.UpdatedAt, foundTopic1.UpdatedAt, 1*time.Second)

	// Verify second topic
	assert.Equal(t, topicID2, foundTopic2.ID)
	assert.Equal(t, diary.ID, foundTopic2.DiaryID)
	assert.Equal(t, "My Second Topic", foundTopic2.Title)
	assert.Equal(t, "Second topic description", foundTopic2.Description)
	assert.Equal(t, "#33FF57", foundTopic2.Color)
	assert.Equal(t, createdTopic2.Version, foundTopic2.Version)
	assert.WithinDuration(t, createdTopic2.CreatedAt, foundTopic2.CreatedAt, 1*time.Second)
	assert.WithinDuration(t, createdTopic2.UpdatedAt, foundTopic2.UpdatedAt, 1*time.Second)
}

func (s *ClientSuite) TestTopic_GetTopics_EmptyList() {
	t := s.T()
	ctx := context.Background()

	// Arrange
	var login = fmt.Sprintf("test-login-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	// Create diary (without topics)
	diary, err := s.client.CreateDiary(ctx, CreateDiaryParams{
		Title:       "Test Diary",
		Description: "Diary for topics testing",
	})
	require.NoError(t, err)

	// Act
	topics, err := s.client.GetTopics(ctx, diary.ID)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, topics)
	require.Empty(t, topics, "Should return empty list when diary has no topics")
}

func (s *ClientSuite) TestTopic_GetTopics_Unauthorized() {
	t := s.T()
	ctx := context.Background()

	// Arrange
	// (no authentication setup)

	// Act
	topics, err := s.client.GetTopics(ctx, "some-diary-id")

	// Assert
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUnauthorized)
	assert.Nil(t, topics)
}

func (s *ClientSuite) TestTopic_GetTopics_DiaryNotFound() {
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
	topics, err := s.client.GetTopics(ctx, nonExistentDiaryID)

	// Assert
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrDiaryNotFound)
	assert.Nil(t, topics)
}
