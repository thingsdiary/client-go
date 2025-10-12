package client

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *ClientSuite) TestTemplate_CreateTemplate() {
	t := s.T()
	ctx := context.Background()

	// Arrange
	var login = fmt.Sprintf("test-login-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	createdDiary, err := s.client.CreateDiary(ctx, CreateDiaryParams{
		Title:       "Test Diary for Template",
		Description: "Test diary for template creation",
	})
	require.NoError(t, err)
	require.NotNil(t, createdDiary)

	// Act
	createdTemplate, err := s.client.CreateTemplate(ctx, createdDiary.ID, CreateTemplateParams{
		Content: "# Daily Journal Template\n\n## Today's Goals\n- \n\n## Reflections\n",
	})

	// Assert
	require.NoError(t, err)
	require.NotNil(t, createdTemplate)

	assert.NotEmpty(t, createdTemplate.ID, "Template ID should be generated")
	assert.Equal(t, createdDiary.ID, createdTemplate.DiaryID)
	assert.Equal(t, "# Daily Journal Template\n\n## Today's Goals\n- \n\n## Reflections\n", createdTemplate.Content)
	assert.Greater(t, createdTemplate.Version, uint64(0))
	assert.WithinDuration(t, time.Now(), createdTemplate.CreatedAt, 5*time.Second)
	assert.WithinDuration(t, time.Now(), createdTemplate.UpdatedAt, 5*time.Second)
}

func (s *ClientSuite) TestTemplate_CreateTemplate_EmptyContent() {
	t := s.T()
	ctx := context.Background()

	// Arrange
	var login = fmt.Sprintf("test-login-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	createdDiary, err := s.client.CreateDiary(ctx, CreateDiaryParams{
		Title:       "Test Diary for Empty Template",
		Description: "Test diary for template with empty content",
	})
	require.NoError(t, err)

	// Act
	createdTemplate, err := s.client.CreateTemplate(ctx, createdDiary.ID, CreateTemplateParams{
		Content: "",
	})

	// Assert
	require.NoError(t, err)
	require.NotNil(t, createdTemplate)

	assert.NotEmpty(t, createdTemplate.ID)
	assert.Equal(t, createdDiary.ID, createdTemplate.DiaryID)
	assert.Equal(t, "", createdTemplate.Content)
}

func (s *ClientSuite) TestTemplate_CreateTemplate_DiaryNotFound() {
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
	template, err := s.client.CreateTemplate(ctx, nonExistentDiaryID, CreateTemplateParams{
		Content: "Test template content",
	})

	// Assert
	require.Error(t, err)
	require.Nil(t, template)
	assert.ErrorIs(t, err, ErrDiaryNotFound)
}

func (s *ClientSuite) TestTemplate_CreateTemplate_Unauthorized() {
	t := s.T()
	ctx := context.Background()

	// Arrange
	// No authentication

	// Act
	template, err := s.client.CreateTemplate(ctx, "some-diary-id", CreateTemplateParams{
		Content: "Test template content",
	})

	// Assert
	require.Error(t, err)
	require.Nil(t, template)
	assert.ErrorIs(t, err, ErrUnauthorized)
}
