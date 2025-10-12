package client

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *ClientSuite) TestTemplate_GetTemplates() {
	t := s.T()
	ctx := context.Background()

	// Arrange
	var login = fmt.Sprintf("test-login-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	// Create diary
	diary, err := s.client.CreateDiary(ctx, CreateDiaryParams{
		Title:       "Test Diary",
		Description: "Diary for templates testing",
	})
	require.NoError(t, err)
	require.NotNil(t, diary)

	// Create first template
	templateID1 := uuid.NewString()
	createdTemplate1, err := s.client.PutTemplate(ctx, diary.ID, templateID1, PutTemplateParams{
		Content: "My first template content with placeholders: {{title}}, {{date}}",
	})
	require.NoError(t, err)
	require.NotNil(t, createdTemplate1)

	// Create second template
	templateID2 := uuid.NewString()
	createdTemplate2, err := s.client.PutTemplate(ctx, diary.ID, templateID2, PutTemplateParams{
		Content: "My second template content with different structure: {{category}}, {{description}}",
	})
	require.NoError(t, err)
	require.NotNil(t, createdTemplate2)

	// Act
	templates, err := s.client.GetTemplates(ctx, diary.ID)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, templates)
	require.Len(t, templates, 2)

	// Find templates by ID
	var foundTemplate1, foundTemplate2 *Template
	for _, template := range templates {
		switch template.ID {
		case templateID1:
			foundTemplate1 = template
		case templateID2:
			foundTemplate2 = template
		}
	}

	require.NotNil(t, foundTemplate1, "First template should be found in the list")
	require.NotNil(t, foundTemplate2, "Second template should be found in the list")

	// Verify first template
	assert.Equal(t, templateID1, foundTemplate1.ID)
	assert.Equal(t, diary.ID, foundTemplate1.DiaryID)
	assert.Equal(t, "My first template content with placeholders: {{title}}, {{date}}", foundTemplate1.Content)
	assert.Equal(t, createdTemplate1.Version, foundTemplate1.Version)
	assert.WithinDuration(t, createdTemplate1.CreatedAt, foundTemplate1.CreatedAt, 1*time.Second)
	assert.WithinDuration(t, createdTemplate1.UpdatedAt, foundTemplate1.UpdatedAt, 1*time.Second)

	// Verify second template
	assert.Equal(t, templateID2, foundTemplate2.ID)
	assert.Equal(t, diary.ID, foundTemplate2.DiaryID)
	assert.Equal(t, "My second template content with different structure: {{category}}, {{description}}", foundTemplate2.Content)
	assert.Equal(t, createdTemplate2.Version, foundTemplate2.Version)
	assert.WithinDuration(t, createdTemplate2.CreatedAt, foundTemplate2.CreatedAt, 1*time.Second)
	assert.WithinDuration(t, createdTemplate2.UpdatedAt, foundTemplate2.UpdatedAt, 1*time.Second)
}

func (s *ClientSuite) TestTemplate_GetTemplates_EmptyList() {
	t := s.T()
	ctx := context.Background()

	// Arrange
	var login = fmt.Sprintf("test-login-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	// Create diary (without templates)
	diary, err := s.client.CreateDiary(ctx, CreateDiaryParams{
		Title:       "Test Diary",
		Description: "Diary for templates testing",
	})
	require.NoError(t, err)

	// Act
	templates, err := s.client.GetTemplates(ctx, diary.ID)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, templates)
	require.Empty(t, templates, "Should return empty list when diary has no templates")
}

func (s *ClientSuite) TestTemplate_GetTemplates_Unauthorized() {
	t := s.T()
	ctx := context.Background()

	// Arrange
	// (no authentication setup)

	// Act
	templates, err := s.client.GetTemplates(ctx, "some-diary-id")

	// Assert
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUnauthorized)
	assert.Nil(t, templates)
}

func (s *ClientSuite) TestTemplate_GetTemplates_DiaryNotFound() {
	t := s.T()
	ctx := context.Background()

	// Arrange
	var login = fmt.Sprintf("test-login-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	nonExistentDiaryID := uuid.NewString()

	// Act
	templates, err := s.client.GetTemplates(ctx, nonExistentDiaryID)

	// Assert
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrDiaryNotFound)
	assert.Nil(t, templates)
}
