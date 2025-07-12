package client

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *ClientSuite) TestTemplate_GetByID_NotFound() {
	t := s.T()
	ctx := context.Background()

	var login = fmt.Sprintf("test-get-template-not-found-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	// Create diary first
	createdDiary, err := s.client.CreateDiary(ctx, CreateDiaryParams{
		Title:       "Test Diary for Template Not Found",
		Description: "Test diary for template not found test",
	})
	require.NoError(t, err)
	require.NotNil(t, createdDiary)

	// Try to get non-existent template
	template, err := s.client.GetTemplateByID(ctx, createdDiary.ID, uuid.NewString())
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrTemplateNotFound)
	assert.Nil(t, template)
}

func (s *ClientSuite) TestTemplate_GetByID_DiaryNotFound() {
	t := s.T()
	ctx := context.Background()

	var login = fmt.Sprintf("test-get-template-diary-not-found-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	// Try to get template from non-existent diary
	template, err := s.client.GetTemplateByID(ctx, uuid.NewString(), uuid.NewString())
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrDiaryNotFound)
	assert.Nil(t, template)
}

func (s *ClientSuite) TestTemplate_GetByID_Unauthorized() {
	t := s.T()
	ctx := context.Background()

	// Try to get template without authentication
	template, err := s.client.GetTemplateByID(ctx, uuid.NewString(), uuid.NewString())
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUnauthorized)
	assert.Nil(t, template)
}

func (s *ClientSuite) TestTemplate_GetByID_AfterCreation() {
	t := s.T()
	ctx := context.Background()

	var login = fmt.Sprintf("test-get-template-after-creation-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	// Create diary first
	createdDiary, err := s.client.CreateDiary(ctx, CreateDiaryParams{
		Title:       "Test Diary for Template Get After Creation",
		Description: "Test diary for template get after creation test",
	})
	require.NoError(t, err)
	require.NotNil(t, createdDiary)

	// Create template
	templateID := uuid.NewString()
	createdTemplate, err := s.client.PutTemplate(ctx, createdDiary.ID, templateID, PutTemplateParams{
		Content: "# Test Template\n\nThis is a test template content.",
	})
	require.NoError(t, err)
	require.NotNil(t, createdTemplate)

	// Get template by ID
	fetchedTemplate, err := s.client.GetTemplateByID(ctx, createdDiary.ID, templateID)
	require.NoError(t, err)
	require.NotNil(t, fetchedTemplate)

	// Assert: Template should match
	assert.Equal(t, createdTemplate.ID, fetchedTemplate.ID)
	assert.Equal(t, createdTemplate.Content, fetchedTemplate.Content)
	assert.Equal(t, createdTemplate.Version, fetchedTemplate.Version)
	assert.WithinDuration(t, createdTemplate.CreatedAt, fetchedTemplate.CreatedAt, time.Second)
	assert.WithinDuration(t, createdTemplate.UpdatedAt, fetchedTemplate.UpdatedAt, time.Second)
}
