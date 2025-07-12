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

func (s *ClientSuite) TestTopic_PutTopic() {
	t := s.T()
	ctx := context.Background()

	// Arrange: Register and authenticate user
	var login = fmt.Sprintf("test-put-topic-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	// Create diary first
	diary, err := s.client.CreateDiary(ctx, CreateDiaryParams{
		Title:       "Test Diary",
		Description: "Diary for topic testing",
	})
	require.NoError(t, err)
	require.NotNil(t, diary)

	// Act: Create topic
	topicID := uuid.NewString()
	createdTopic, err := s.client.PutTopic(ctx, diary.ID, topicID, PutTopicParams{
		Title:       "Work Projects",
		Description: "Topics related to work projects",
		Color:       "#FF5733",
	})

	// Assert: Topic creation was successful
	require.NoError(t, err)
	require.NotNil(t, createdTopic)
	assert.Equal(t, topicID, createdTopic.ID)
	assert.Equal(t, diary.ID, createdTopic.DiaryID)
	assert.Equal(t, "Work Projects", createdTopic.Title)
	assert.Equal(t, "Topics related to work projects", createdTopic.Description)
	assert.Equal(t, "#FF5733", createdTopic.Color)
	assert.Greater(t, createdTopic.Version, uint64(0))

	// Act: Update topic
	updatedTopic, err := s.client.PutTopic(ctx, diary.ID, topicID, PutTopicParams{
		Title:       "Updated Work Projects",
		Description: "Updated description for work projects",
		Color:       "#33FF57",
	})

	// Assert: Topic update was successful
	require.NoError(t, err)
	require.NotNil(t, updatedTopic)
	assert.Equal(t, topicID, updatedTopic.ID)
	assert.Equal(t, diary.ID, updatedTopic.DiaryID)
	assert.Equal(t, "Updated Work Projects", updatedTopic.Title)
	assert.Equal(t, "Updated description for work projects", updatedTopic.Description)
	assert.Equal(t, "#33FF57", updatedTopic.Color)
	assert.Greater(t, updatedTopic.Version, createdTopic.Version)
}

func (s *ClientSuite) TestTopic_PutTopic_EmptyFields() {
	t := s.T()
	ctx := context.Background()

	// Arrange: Register and authenticate user
	var login = fmt.Sprintf("test-put-topic-empty-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	// Create diary first
	diary, err := s.client.CreateDiary(ctx, CreateDiaryParams{
		Title:       "Test Diary",
		Description: "Diary for topic testing",
	})
	require.NoError(t, err)
	require.NotNil(t, diary)

	// Act: Create topic with empty fields
	topicID := uuid.NewString()
	createdTopic, err := s.client.PutTopic(ctx, diary.ID, topicID, PutTopicParams{
		Title:       "",
		Description: "",
		Color:       "",
	})

	// Assert: Topic creation with empty fields was successful
	require.NoError(t, err)
	require.NotNil(t, createdTopic)
	assert.Equal(t, topicID, createdTopic.ID)
	assert.Equal(t, diary.ID, createdTopic.DiaryID)
	assert.Equal(t, "", createdTopic.Title)
	assert.Equal(t, "", createdTopic.Description)
	assert.Equal(t, "", createdTopic.Color)
}

func (s *ClientSuite) TestTopic_PutTopic_NotFound() {
	t := s.T()
	ctx := context.Background()

	// Arrange: Register and authenticate user
	var login = fmt.Sprintf("test-put-topic-not-found-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	// Act: Try to create topic in non-existent diary
	topicID := uuid.NewString()
	topic, err := s.client.PutTopic(ctx, uuid.NewString(), topicID, PutTopicParams{
		Title:       "Test Topic",
		Description: "Test description",
		Color:       "#FF5733",
	})

	// Assert: Should get diary not found error
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrDiaryNotFound)
	assert.Nil(t, topic)
}

func (s *ClientSuite) TestTopic_PutTopic_Unauthorized() {
	t := s.T()
	ctx := context.Background()

	// Act: Try to create topic without authentication
	topicID := uuid.NewString()
	topic, err := s.client.PutTopic(ctx, uuid.NewString(), topicID, PutTopicParams{
		Title:       "Test Topic",
		Description: "Test description",
		Color:       "#FF5733",
	})

	// Assert: Should get unauthorized error
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUnauthorized)
	assert.Nil(t, topic)
}

func (s *ClientSuite) TestTopic_PutTopicWithDefaultTemplate() {
	t := s.T()
	ctx := context.Background()

	// Arrange: Register and authenticate user
	var login = fmt.Sprintf("test-put-topic-template-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	// Create diary first
	diary, err := s.client.CreateDiary(ctx, CreateDiaryParams{
		Title:       "Test Diary",
		Description: "Diary for topic with template testing",
	})
	require.NoError(t, err)
	require.NotNil(t, diary)

	// Create template first
	template, err := s.client.CreateTemplate(ctx, diary.ID, CreateTemplateParams{
		Content: "## Project: \n\n### Tasks:\n- [ ] \n\n### Notes:\n",
	})
	require.NoError(t, err)
	templateID := template.ID

	// Act: Create topic with default template
	createdTopic, err := s.client.CreateTopic(ctx, diary.ID, CreateTopicParams{
		Title:             "Work Projects",
		Description:       "Topics related to work projects",
		Color:             "#FF5733",
		DefaultTemplateID: mo.Some(templateID),
	})
	topicID := createdTopic.ID

	// Assert: Topic creation was successful with template
	require.NoError(t, err)
	require.NotNil(t, createdTopic)
	assert.Equal(t, topicID, createdTopic.ID)
	assert.Equal(t, diary.ID, createdTopic.DiaryID)
	assert.Equal(t, "Work Projects", createdTopic.Title)
	assert.Equal(t, "Topics related to work projects", createdTopic.Description)
	assert.Equal(t, "#FF5733", createdTopic.Color)
	assert.True(t, createdTopic.DefaultTemplateID.IsPresent())
	assert.Equal(t, templateID, createdTopic.DefaultTemplateID.MustGet())
	assert.Greater(t, createdTopic.Version, uint64(0))

	// Act: Remove default template
	updatedTopic, err := s.client.PutTopic(ctx, diary.ID, topicID, PutTopicParams{
		Title:             "Work Projects",
		Description:       "Topics related to work projects",
		Color:             "#FF5733",
		DefaultTemplateID: mo.None[string](),
	})

	// Assert: Template was removed
	require.NoError(t, err)
	require.NotNil(t, updatedTopic)
	assert.False(t, updatedTopic.DefaultTemplateID.IsPresent())
}
