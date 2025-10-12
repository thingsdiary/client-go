package client

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *ClientSuite) TestTemplate_PutTemplate() {
	t := s.T()
	ctx := context.Background()

	// Arrange: Register and authenticate user
	var login = fmt.Sprintf("test-put-template-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	// Create diary first
	diary, err := s.client.CreateDiary(ctx, CreateDiaryParams{
		Title:       "Test Diary",
		Description: "Diary for template testing",
	})
	require.NoError(t, err)

	// Act: Create template
	templateID := uuid.NewString()
	createdTemplate, err := s.client.PutTemplate(ctx, diary.ID, templateID, PutTemplateParams{
		Content: "# Daily Journal\n\n## Today's Goals\n- \n\n## Reflections\n",
	})

	// Assert: Template created successfully
	require.NoError(t, err)
	require.NotNil(t, createdTemplate)
	assert.Equal(t, templateID, createdTemplate.ID)
	assert.Equal(t, "# Daily Journal\n\n## Today's Goals\n- \n\n## Reflections\n", createdTemplate.Content)
	assert.WithinDuration(t, time.Now(), createdTemplate.CreatedAt, time.Second)
	assert.WithinDuration(t, time.Now(), createdTemplate.UpdatedAt, time.Second)

	// Act: Update template
	updatedTemplate, err := s.client.PutTemplate(ctx, diary.ID, templateID, PutTemplateParams{
		Content: "# Daily Journal\n\n## Goals\n- \n\n## What I learned\n\n## Gratitude\n",
	})

	// Assert: Template updated successfully
	require.NoError(t, err)
	require.NotNil(t, updatedTemplate)
	assert.Equal(t, templateID, updatedTemplate.ID)
	assert.Equal(t, "# Daily Journal\n\n## Goals\n- \n\n## What I learned\n\n## Gratitude\n", updatedTemplate.Content)
	assert.Greater(t, updatedTemplate.Version, createdTemplate.Version)
	assert.WithinDuration(t, createdTemplate.CreatedAt, updatedTemplate.CreatedAt, time.Second)
	assert.True(t, updatedTemplate.UpdatedAt.After(createdTemplate.UpdatedAt))
}

func (s *ClientSuite) TestTemplate_PutTemplate_NotFound() {
	t := s.T()
	ctx := context.Background()

	// Arrange: Register and authenticate user
	var login = fmt.Sprintf("test-put-template-not-found-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	// Act: Try to create template in non-existent diary
	templateID := uuid.NewString()
	template, err := s.client.PutTemplate(ctx, uuid.NewString(), templateID, PutTemplateParams{
		Content: "Test content",
	})

	// Assert: Should get diary not found error
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrDiaryNotFound)
	assert.Nil(t, template)
}

func (s *ClientSuite) TestTemplate_PutTemplate_Unauthorized() {
	t := s.T()
	ctx := context.Background()

	// Act: Try to create template without authentication
	templateID := uuid.NewString()
	template, err := s.client.PutTemplate(ctx, uuid.NewString(), templateID, PutTemplateParams{
		Content: "Test content",
	})

	// Assert: Should get unauthorized error
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUnauthorized)
	assert.Nil(t, template)
}

func (s *ClientSuite) TestTemplate_PutTemplate_EmptyFields() {
	t := s.T()
	ctx := context.Background()

	// Arrange: Register and authenticate user
	var login = fmt.Sprintf("test-put-template-empty-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	// Create diary first
	diary, err := s.client.CreateDiary(ctx, CreateDiaryParams{
		Title:       "Test Diary",
		Description: "Diary for template testing",
	})
	require.NoError(t, err)

	// Act: Create template with empty fields
	templateID := uuid.NewString()
	createdTemplate, err := s.client.PutTemplate(ctx, diary.ID, templateID, PutTemplateParams{
		Content: "",
	})

	// Assert: Template created successfully with empty fields
	require.NoError(t, err)
	require.NotNil(t, createdTemplate)
	assert.Equal(t, templateID, createdTemplate.ID)
	assert.Equal(t, "", createdTemplate.Content)
}
