package client

import (
	"context"
	"fmt"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *ClientSuite) TestDiary_Create() {
	t := s.T()
	ctx := context.Background()

	var login = fmt.Sprintf("test-login-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	diary, err := s.client.CreateDiary(ctx, CreateDiaryParams{
		Title:       "My Personal Diary",
		Description: "Here will be my secret entries",
	})
	require.NoError(t, err)
	require.NotNil(t, diary)

	assert.NotEmpty(t, diary.ID)
	assert.Equal(t, "My Personal Diary", diary.Title)
	assert.Equal(t, "Here will be my secret entries", diary.Description)
}

func (s *ClientSuite) TestDiary_Create_LimitExceeded() {
	t := s.T()
	ctx := context.Background()

	var login = fmt.Sprintf("test-limit-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	// Create 10 diaries to reach the limit
	for i := 0; i < 11; i++ {
		_, err := s.client.CreateDiary(ctx, CreateDiaryParams{
			Title:       fmt.Sprintf("Diary %d", i+1),
			Description: fmt.Sprintf("Description for diary %d", i+1),
		})
		require.NoError(t, err, "Failed to create diary %d", i+1)
	}

	// Try to create 11th diary - should fail with limit exceeded
	_, err = s.client.CreateDiary(ctx, CreateDiaryParams{
		Title:       "Diary 11",
		Description: "This should fail due to limit",
	})

	// Assert: Should get diary limit exceeded error
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrDiaryLimitExceeded)
}
