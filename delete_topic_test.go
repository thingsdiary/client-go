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

func (s *ClientSuite) TestTopic_DeleteTopic_ThenUpdateDeleted() {
	t := s.T()
	ctx := context.Background()

	// Arrange: Register and authenticate user
	var login = fmt.Sprintf("test-delete-topic-then-update-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	// Create diary
	createdDiary, err := s.client.CreateDiary(ctx, CreateDiaryParams{
		Title:       "Test Diary",
		Description: "Diary for delete then update test",
	})
	require.NoError(t, err)
	require.NotNil(t, createdDiary)

	// Create topic
	createdTopic, err := s.client.CreateTopic(ctx, createdDiary.ID, CreateTopicParams{
		Title:       "Topic to delete and update",
		Description: "This topic will be deleted then updated",
		Color:       "#FF0000",
	})
	require.NoError(t, err)
	require.NotNil(t, createdTopic)

	// Delete topic
	err = s.client.DeleteTopic(ctx, createdDiary.ID, createdTopic.ID)
	require.NoError(t, err)

	// Act: Try to update deleted topic
	gotTopic, err := s.client.PutTopic(ctx, createdDiary.ID, createdTopic.ID, PutTopicParams{
		Title:       "Updated Title",
		Description: "Updated Description",
		Color:       "#0000FF",
	})
	require.NoError(t, err)

	// Assert
	assert.Equal(t, gotTopic.Title, "Updated Title")
}

func (s *ClientSuite) TestTopic_DeleteTopic_EntriesPreserved() {
	t := s.T()
	ctx := context.Background()

	// Arrange: Register and authenticate user
	var login = fmt.Sprintf("test-delete-topic-entries-preserved-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	// Create diary
	createdDiary, err := s.client.CreateDiary(ctx, CreateDiaryParams{
		Title:       "Test Diary",
		Description: "Diary for entries preserved test",
	})
	require.NoError(t, err)
	require.NotNil(t, createdDiary)

	// Create topic
	createdTopic, err := s.client.CreateTopic(ctx, createdDiary.ID, CreateTopicParams{
		Title:       "Topic with entries",
		Description: "This topic has entries",
		Color:       "#FF0000",
	})
	require.NoError(t, err)
	require.NotNil(t, createdTopic)

	// Create entries in the topic
	entry1, err := s.client.CreateEntry(ctx, createdDiary.ID, CreateEntryParams{
		Content: "Entry 1 in topic",
		TopicID: mo.Some(createdTopic.ID),
	})
	require.NoError(t, err)
	require.NotNil(t, entry1)

	entry2, err := s.client.CreateEntry(ctx, createdDiary.ID, CreateEntryParams{
		Content: "Entry 2 in topic",
		TopicID: mo.Some(createdTopic.ID),
	})
	require.NoError(t, err)
	require.NotNil(t, entry2)

	// Act: Delete topic without params (DeleteEntries defaults to false)
	err = s.client.DeleteTopic(ctx, createdDiary.ID, createdTopic.ID)
	require.NoError(t, err)

	// Assert: Topic is deleted
	gotTopic, err := s.client.GetTopicByID(ctx, createdDiary.ID, createdTopic.ID)
	require.NoError(t, err)
	assert.True(t, gotTopic.DeletedAt.IsPresent())

	// Assert: Entries are preserved but unlinked from topic
	gotEntry1, err := s.client.GetEntryByID(ctx, createdDiary.ID, entry1.ID)
	require.NoError(t, err)
	assert.False(t, gotEntry1.DeletedAt.IsPresent(), "Entry 1 should not be deleted")
	assert.False(t, gotEntry1.TopicID.IsPresent(), "Entry 1 should be unlinked from topic")

	gotEntry2, err := s.client.GetEntryByID(ctx, createdDiary.ID, entry2.ID)
	require.NoError(t, err)
	assert.False(t, gotEntry2.DeletedAt.IsPresent(), "Entry 2 should not be deleted")
	assert.False(t, gotEntry2.TopicID.IsPresent(), "Entry 2 should be unlinked from topic")
}

func (s *ClientSuite) TestTopic_DeleteTopic_WithEntries() {
	t := s.T()
	ctx := context.Background()

	// Arrange: Register and authenticate user
	var login = fmt.Sprintf("test-delete-topic-with-entries-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	// Create diary
	createdDiary, err := s.client.CreateDiary(ctx, CreateDiaryParams{
		Title:       "Test Diary",
		Description: "Diary for entries deletion test",
	})
	require.NoError(t, err)
	require.NotNil(t, createdDiary)

	// Create topic
	createdTopic, err := s.client.CreateTopic(ctx, createdDiary.ID, CreateTopicParams{
		Title:       "Topic with entries to delete",
		Description: "This topic and its entries will be deleted",
		Color:       "#FF0000",
	})
	require.NoError(t, err)
	require.NotNil(t, createdTopic)

	// Create entries in the topic
	entry1, err := s.client.CreateEntry(ctx, createdDiary.ID, CreateEntryParams{
		Content: "Entry 1 to delete",
		TopicID: mo.Some(createdTopic.ID),
	})
	require.NoError(t, err)
	require.NotNil(t, entry1)

	entry2, err := s.client.CreateEntry(ctx, createdDiary.ID, CreateEntryParams{
		Content: "Entry 2 to delete",
		TopicID: mo.Some(createdTopic.ID),
	})
	require.NoError(t, err)
	require.NotNil(t, entry2)

	// Create entry without topic (should not be deleted)
	entry3, err := s.client.CreateEntry(ctx, createdDiary.ID, CreateEntryParams{
		Content: "Entry 3 without topic",
		TopicID: mo.None[string](),
	})
	require.NoError(t, err)
	require.NotNil(t, entry3)

	// Act: Delete topic with DeleteEntries=true
	err = s.client.DeleteTopic(ctx, createdDiary.ID, createdTopic.ID, DeleteTopicParams{
		DeleteEntries: true,
	})
	require.NoError(t, err)

	// Assert: Topic is deleted
	gotTopic, err := s.client.GetTopicByID(ctx, createdDiary.ID, createdTopic.ID)
	require.NoError(t, err)
	assert.True(t, gotTopic.DeletedAt.IsPresent())

	// Assert: Entries in topic are deleted
	gotEntry1, err := s.client.GetEntryByID(ctx, createdDiary.ID, entry1.ID)
	require.NoError(t, err)
	assert.True(t, gotEntry1.DeletedAt.IsPresent(), "Entry 1 should be deleted")

	gotEntry2, err := s.client.GetEntryByID(ctx, createdDiary.ID, entry2.ID)
	require.NoError(t, err)
	assert.True(t, gotEntry2.DeletedAt.IsPresent(), "Entry 2 should be deleted")

	// Assert: Entry without topic is not deleted
	gotEntry3, err := s.client.GetEntryByID(ctx, createdDiary.ID, entry3.ID)
	require.NoError(t, err)
	assert.False(t, gotEntry3.DeletedAt.IsPresent(), "Entry 3 should not be deleted")
}
