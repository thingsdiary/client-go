package client

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func (s *ClientSuite) TestLogin_Authenticate() {
	t := s.T()
	ctx := context.Background()

	var login = fmt.Sprintf("test-login-%d@thingsdiary.io", time.Now().UnixMilli())

	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	// Verify user is authenticated
	require.NotEmpty(t, s.client.authToken)
	require.NotNil(t, s.client.credentials)
}

func (s *ClientSuite) TestLogin_AuthenticateWithInvalidCredentials() {
	t := s.T()
	ctx := context.Background()

	t.Run("wrong credentials", func(t *testing.T) {
		var login = fmt.Sprintf("test-invalid-login-%d@thingsdiary.io", time.Now().UnixMilli())

		// Register user with valid credentials
		err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
		require.NoError(t, err)

		// Test with wrong password
		err = s.client.Authenticate(ctx, login, "wrong-password", s.seedPhrase)
		require.ErrorIs(t, err, ErrInvalidCredentials)
		require.Empty(t, s.client.authToken)
		require.Nil(t, s.client.credentials)
	})

	t.Run("invalid credentials", func(t *testing.T) {
		// Test with non-existent user
		err := s.client.Authenticate(ctx, "non-existent@thingsdiary.io", "password-123", s.seedPhrase)
		require.ErrorIs(t, err, ErrInvalidCredentials)
		require.Empty(t, s.client.authToken)
		require.Nil(t, s.client.credentials)
	})
}
