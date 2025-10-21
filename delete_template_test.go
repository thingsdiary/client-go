package client

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *ClientSuite) TestTemplate_DeleteTemplate() {
	t := s.T()
	ctx := context.Background()

	// Arrange: Register and authenticate user
	var login = fmt.Sprintf("test-delete-template-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	// Create diary
	createdDiary, err := s.client.CreateDiary(ctx, CreateDiaryParams{
		Title:       "Test Diary",
		Description: "Diary for template deletion test",
	})
	require.NoError(t, err)
	require.NotNil(t, createdDiary)

	// Create template to delete
	createdTemplate, err := s.client.CreateTemplate(ctx, createdDiary.ID, CreateTemplateParams{
		Content: "Template to be deleted",
	})
	require.NoError(t, err)
	require.NotNil(t, createdTemplate)

	// Act: Delete template
	err = s.client.DeleteTemplate(ctx, createdDiary.ID, createdTemplate.ID)
	require.NoError(t, err)

	// Assert
	gotTemplate, err := s.client.GetTemplateByID(ctx, createdDiary.ID, createdTemplate.ID)
	require.NoError(t, err)
	assert.True(t, gotTemplate.DeletedAt.IsPresent())
}

func (s *ClientSuite) TestTemplate_DeleteTemplate_NotFound() {
	t := s.T()
	ctx := context.Background()

	// Arrange: Register and authenticate user
	var login = fmt.Sprintf("test-delete-template-not-found-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	// Create diary
	createdDiary, err := s.client.CreateDiary(ctx, CreateDiaryParams{
		Title:       "Test Diary",
		Description: "Diary for template not found test",
	})
	require.NoError(t, err)
	require.NotNil(t, createdDiary)

	// Act: Try to delete non-existent template
	err = s.client.DeleteTemplate(ctx, createdDiary.ID, uuid.NewString())

	// Assert: Returns template not found error
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrTemplateNotFound)
}

func (s *ClientSuite) TestTemplate_DeleteTemplate_Unauthorized() {
	t := s.T()
	ctx := context.Background()

	// Act: Try to delete template without authentication
	err := s.client.DeleteTemplate(ctx, uuid.NewString(), uuid.NewString())

	// Assert: Returns unauthorized error
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUnauthorized)
}
