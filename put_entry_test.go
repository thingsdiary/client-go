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

func (s *ClientSuite) TestEntry_PutEntry() {
	t := s.T()
	ctx := context.Background()

	// Arrange: Register and authenticate user
	var login = fmt.Sprintf("test-put-entry-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	// Create diary first
	diary, err := s.client.CreateDiary(ctx, CreateDiaryParams{
		Title:       "Test Diary",
		Description: "Diary for entry testing",
	})
	require.NoError(t, err)
	require.NotNil(t, diary)

	// Act: Create entry
	entryID := uuid.NewString()
	createdEntry, err := s.client.PutEntry(ctx, diary.ID, entryID, PutEntryParams{
		Content:       "My first entry content",
		TopicID:       mo.None[string](),
		Archived:      false,
		Bookmarked:    true,
		PreviewHidden: false,
	})

	// Assert: Entry creation was successful
	require.NoError(t, err)
	require.NotNil(t, createdEntry)
	assert.Equal(t, entryID, createdEntry.ID)
	assert.Equal(t, "My first entry content", createdEntry.Content)
	assert.False(t, createdEntry.TopicID.IsPresent())
	assert.False(t, createdEntry.Archived)
	assert.True(t, createdEntry.Bookmarked)
	assert.False(t, createdEntry.PreviewHidden)
	assert.False(t, createdEntry.DeletedAt.IsPresent())
	assert.Greater(t, createdEntry.Version, uint64(0))

	// Act: Update entry
	updatedEntry, err := s.client.PutEntry(ctx, diary.ID, entryID, PutEntryParams{
		Content:       "Updated entry content",
		TopicID:       mo.None[string](),
		Archived:      true,
		Bookmarked:    false,
		PreviewHidden: true,
	})

	// Assert: Entry update was successful
	require.NoError(t, err)
	require.NotNil(t, updatedEntry)
	assert.Equal(t, entryID, updatedEntry.ID)
	assert.Equal(t, "Updated entry content", updatedEntry.Content)
	assert.True(t, updatedEntry.Archived)
	assert.False(t, updatedEntry.Bookmarked)
	assert.True(t, updatedEntry.PreviewHidden)
	assert.Greater(t, updatedEntry.Version, createdEntry.Version)
}

func (s *ClientSuite) TestEntry_PutEntry_WithTopic() {
	t := s.T()
	ctx := context.Background()

	// Arrange: Register and authenticate user
	var login = fmt.Sprintf("test-put-entry-topic-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	// Create diary
	diary, err := s.client.CreateDiary(ctx, CreateDiaryParams{
		Title:       "Test Diary",
		Description: "Diary for entry testing",
	})
	require.NoError(t, err)

	// Create topic first
	topicID := uuid.NewString()
	_, err = s.client.PutTopic(ctx, diary.ID, topicID, PutTopicParams{
		Title:       "Test Topic",
		Description: "Topic for entry testing",
		Color:       "#FF5733",
	})
	require.NoError(t, err)

	// Act: Create entry with topic
	entryID := uuid.NewString()
	createdEntry, err := s.client.PutEntry(ctx, diary.ID, entryID, PutEntryParams{
		Content:       "Entry with topic",
		TopicID:       mo.Some(topicID),
		Archived:      false,
		Bookmarked:    false,
		PreviewHidden: false,
	})

	// Assert: Entry with topic was created successfully
	require.NoError(t, err)
	require.NotNil(t, createdEntry)
	assert.Equal(t, entryID, createdEntry.ID)
	assert.Equal(t, "Entry with topic", createdEntry.Content)
	assert.True(t, createdEntry.TopicID.IsPresent())
	assert.Equal(t, topicID, createdEntry.TopicID.MustGet())
}

func (s *ClientSuite) TestEntry_PutEntry_NotFound() {
	t := s.T()
	ctx := context.Background()

	// Arrange: Register and authenticate user
	var login = fmt.Sprintf("test-put-entry-not-found-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	// Act: Try to create entry in non-existent diary
	entryID := uuid.NewString()
	nonExistentDiaryID := uuid.NewString()
	entry, err := s.client.PutEntry(ctx, nonExistentDiaryID, entryID, PutEntryParams{
		Content: "Test content",
	})

	// Assert: Returns diary not found error
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrDiaryNotFound)
	assert.Nil(t, entry)
}

func (s *ClientSuite) TestEntry_PutEntry_Unauthorized() {
	t := s.T()
	ctx := context.Background()

	// Act: Try to create entry without authentication
	entryID := uuid.NewString()
	diaryID := uuid.NewString()
	entry, err := s.client.PutEntry(ctx, diaryID, entryID, PutEntryParams{
		Content: "Test content",
	})

	// Assert: Returns unauthorized error
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUnauthorized)
	assert.Nil(t, entry)
}

func (s *ClientSuite) TestEntry_PutEntry_EmptyContent() {
	t := s.T()
	ctx := context.Background()

	// Arrange: Register and authenticate user
	var login = fmt.Sprintf("test-put-entry-empty-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	// Create diary
	diary, err := s.client.CreateDiary(ctx, CreateDiaryParams{
		Title:       "Test Diary",
		Description: "Diary for entry testing",
	})
	require.NoError(t, err)

	// Act: Create entry with empty content
	entryID := uuid.NewString()
	createdEntry, err := s.client.PutEntry(ctx, diary.ID, entryID, PutEntryParams{
		Content:       "",
		TopicID:       mo.None[string](),
		Archived:      false,
		Bookmarked:    false,
		PreviewHidden: false,
	})

	// Assert: Entry with empty content is allowed
	require.NoError(t, err)
	require.NotNil(t, createdEntry)
	assert.Equal(t, entryID, createdEntry.ID)
	assert.Equal(t, "", createdEntry.Content)
}

func (s *ClientSuite) TestEntry_PutEntry_TopicNotFound() {
	t := s.T()
	ctx := context.Background()

	// Arrange: Register and authenticate user
	var login = fmt.Sprintf("test-put-entry-topic-not-found-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	// Create diary first
	diary, err := s.client.CreateDiary(ctx, CreateDiaryParams{
		Title:       "Test Diary",
		Description: "Diary for entry testing",
	})
	require.NoError(t, err)

	// Act: Try to create entry with non-existent topic_id
	entryID := uuid.NewString()
	nonExistentTopicID := uuid.NewString()

	_, err = s.client.PutEntry(ctx, diary.ID, entryID, PutEntryParams{
		Content: "Test entry content",
		TopicID: mo.Some(nonExistentTopicID),
	})

	// Assert: Should get topic not found error
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrTopicNotFound)
}
