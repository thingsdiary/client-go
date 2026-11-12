package client

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "github.com/thingsdiary/client-go/proto"
)

func TestNextEntryStateChange_BuildsAppendChange(t *testing.T) {
	t.Parallel()

	// arrange
	current := &v1.EncryptedEntry{
		Change: &v1.EncryptedChange{
			Change:     &v1.Change{Id: "parent", Payload: []byte("payload")},
			Encryption: &v1.Encryption{DiaryKeyId: "key"},
		},
	}

	// act
	change, err := nextEntryStateChange(current, "entry")

	// assert
	require.NoError(t, err)
	require.NotEmpty(t, change.GetChange().GetId())
	require.Equal(t, "entry", change.GetChange().GetObjectId())
	require.Equal(t, []string{"parent"}, change.GetChange().GetParentChangeIds())
	require.Equal(t, []byte("payload"), change.GetChange().GetPayload())
	require.Equal(t, "key", change.GetEncryption().GetDiaryKeyId())
}

func (s *ClientSuite) TestEntry_DeleteEntry() {
	t := s.T()
	ctx := context.Background()

	// Arrange: Register and authenticate user
	var login = fmt.Sprintf("test-delete-entry-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase, s.kdfSalt)
	require.NoError(t, err)

	err = s.authenticate(ctx, login, "password-123")
	require.NoError(t, err)

	// Create diary
	createdDiary, err := s.client.CreateDiary(ctx, CreateDiaryParams{
		Title:       "Test Diary",
		Description: "Diary for entry deletion test",
	})
	require.NoError(t, err)
	require.NotNil(t, createdDiary)

	createdEntry, err := s.client.CreateEntry(ctx, createdDiary.ID, CreateEntryParams{
		Content: "Entry to be deleted",
	})
	require.NoError(t, err)
	require.NotNil(t, createdEntry)

	// Act: Delete entry
	err = s.client.DeleteEntry(ctx, createdDiary.ID, createdEntry.ID)
	require.NoError(t, err)

	// Assert
	gotEntry, err := s.client.GetEntryByID(ctx, createdDiary.ID, createdEntry.ID)
	require.NoError(t, err)
	assert.True(t, gotEntry.DeletedAt.IsPresent())
}

func (s *ClientSuite) TestEntry_DeleteEntry_NotFound() {
	t := s.T()
	ctx := context.Background()

	// Arrange: Register and authenticate user
	var login = fmt.Sprintf("test-delete-entry-not-found-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase, s.kdfSalt)
	require.NoError(t, err)

	err = s.authenticate(ctx, login, "password-123")
	require.NoError(t, err)

	// Create diary
	createdDiary, err := s.client.CreateDiary(ctx, CreateDiaryParams{
		Title:       "Test Diary",
		Description: "Diary for entry not found test",
	})
	require.NoError(t, err)
	require.NotNil(t, createdDiary)

	// Act: Try to delete non-existent entry
	err = s.client.DeleteEntry(ctx, createdDiary.ID, uuid.NewString())

	// Assert: Returns entry not found error
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrEntryNotFound)
}

func (s *ClientSuite) TestEntry_DeleteEntry_Unauthorized() {
	t := s.T()
	ctx := context.Background()

	// Act: Try to delete entry without authentication
	err := s.client.DeleteEntry(ctx, uuid.NewString(), uuid.NewString())

	// Assert: Returns unauthorized error
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUnauthorized)
}
