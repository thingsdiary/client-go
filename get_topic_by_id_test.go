package client

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *ClientSuite) TestTopic_GetByID() {
	t := s.T()
	ctx := context.Background()

	var login = fmt.Sprintf("test-get-topic-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	// Create diary first
	createdDiary, err := s.client.CreateDiary(ctx, CreateDiaryParams{
		Title:       "Test Diary for Topic",
		Description: "Test diary for topic retrieval",
	})
	require.NoError(t, err)
	require.NotNil(t, createdDiary)

	// Create topic
	topicID := uuid.NewString()
	createdTopic, err := s.client.PutTopic(ctx, createdDiary.ID, topicID, PutTopicParams{
		Title:       "Personal Notes",
		Description: "Topics for personal notes and thoughts",
		Color:       "#3366FF",
	})
	require.NoError(t, err)
	require.NotNil(t, createdTopic)

	// Get topic by ID
	fetchedTopic, err := s.client.GetTopicByID(ctx, createdDiary.ID, topicID)
	require.NoError(t, err)
	require.NotNil(t, fetchedTopic)

	// Verify topic content
	assert.Equal(t, topicID, fetchedTopic.ID)
	assert.Equal(t, createdDiary.ID, fetchedTopic.DiaryID)
	assert.Equal(t, "Personal Notes", fetchedTopic.Title)
	assert.Equal(t, "Topics for personal notes and thoughts", fetchedTopic.Description)
	assert.Equal(t, "#3366FF", fetchedTopic.Color)
	assert.Equal(t, createdTopic.Version, fetchedTopic.Version)
	assert.WithinDuration(t, createdTopic.CreatedAt, fetchedTopic.CreatedAt, time.Second)
	assert.WithinDuration(t, createdTopic.UpdatedAt, fetchedTopic.UpdatedAt, time.Second)
}

func (s *ClientSuite) TestTopic_GetByID_NotFound() {
	t := s.T()
	ctx := context.Background()

	var login = fmt.Sprintf("test-get-topic-not-found-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	// Create diary first
	createdDiary, err := s.client.CreateDiary(ctx, CreateDiaryParams{
		Title:       "Test Diary for Topic Not Found",
		Description: "Test diary for topic not found test",
	})
	require.NoError(t, err)
	require.NotNil(t, createdDiary)

	// Try to get non-existent topic
	topic, err := s.client.GetTopicByID(ctx, createdDiary.ID, uuid.NewString())
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrTopicNotFound)
	assert.Nil(t, topic)
}

func (s *ClientSuite) TestTopic_GetByID_DiaryNotFound() {
	t := s.T()
	ctx := context.Background()

	var login = fmt.Sprintf("test-get-topic-diary-not-found-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	// Try to get topic from non-existent diary
	topic, err := s.client.GetTopicByID(ctx, uuid.NewString(), uuid.NewString())
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrDiaryNotFound)
	assert.Nil(t, topic)
}

func (s *ClientSuite) TestTopic_GetByID_Unauthorized() {
	t := s.T()
	ctx := context.Background()

	// Try to get topic without authentication
	topic, err := s.client.GetTopicByID(ctx, uuid.NewString(), uuid.NewString())
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUnauthorized)
	assert.Nil(t, topic)
}

func (s *ClientSuite) TestTopic_GetByID_EmptyFields() {
	t := s.T()
	ctx := context.Background()

	var login = fmt.Sprintf("test-get-topic-empty-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	// Create diary
	createdDiary, err := s.client.CreateDiary(ctx, CreateDiaryParams{
		Title:       "Test Diary for Empty Topic",
		Description: "Test diary for topic with empty fields",
	})
	require.NoError(t, err)
	require.NotNil(t, createdDiary)

	// Create topic with empty fields
	topicID := uuid.NewString()
	createdTopic, err := s.client.PutTopic(ctx, createdDiary.ID, topicID, PutTopicParams{
		Title:       "",
		Description: "",
		Color:       "",
	})
	require.NoError(t, err)
	require.NotNil(t, createdTopic)

	// Get topic by ID
	fetchedTopic, err := s.client.GetTopicByID(ctx, createdDiary.ID, topicID)
	require.NoError(t, err)
	require.NotNil(t, fetchedTopic)

	// Verify topic content with empty fields
	assert.Equal(t, topicID, fetchedTopic.ID)
	assert.Equal(t, createdDiary.ID, fetchedTopic.DiaryID)
	assert.Equal(t, "", fetchedTopic.Title)
	assert.Equal(t, "", fetchedTopic.Description)
	assert.Equal(t, "", fetchedTopic.Color)
	assert.Equal(t, createdTopic.Version, fetchedTopic.Version)
}
