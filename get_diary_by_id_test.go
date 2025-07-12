package client

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *ClientSuite) TestDiary_GetByID() {
	t := s.T()
	ctx := context.Background()

	var login = fmt.Sprintf("test-login-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	createdDiary, err := s.client.CreateDiary(ctx, CreateDiaryParams{
		Title:       "My Test Diary",
		Description: "Test diary for GetByID test",
	})
	require.NoError(t, err)
	require.NotNil(t, createdDiary)

	fetchedDiary, err := s.client.GetDiaryByID(ctx, createdDiary.ID)
	require.NoError(t, err)
	require.NotNil(t, fetchedDiary)

	assert.Equal(t, createdDiary.ID, fetchedDiary.ID)
	assert.Equal(t, "My Test Diary", fetchedDiary.Title)
	assert.Equal(t, "Test diary for GetByID test", fetchedDiary.Description)
	assert.Equal(t, createdDiary.Version, fetchedDiary.Version)
	assert.WithinDuration(t, createdDiary.CreatedAt, fetchedDiary.CreatedAt, time.Second)
	assert.WithinDuration(t, createdDiary.UpdatedAt, fetchedDiary.UpdatedAt, time.Second)
}

func (s *ClientSuite) TestDiary_GetByID_NotFound() {
	t := s.T()
	ctx := context.Background()

	var login = fmt.Sprintf("test-login-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	diary, err := s.client.GetDiaryByID(ctx, uuid.NewString())
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrDiaryNotFound)
	assert.Nil(t, diary)
}

func (s *ClientSuite) TestDiary_GetByID_Unauthorized() {
	t := s.T()
	ctx := context.Background()

	diary, err := s.client.GetDiaryByID(ctx, uuid.NewString())
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUnauthorized)
	assert.Nil(t, diary)
}
