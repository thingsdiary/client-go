package client

import (
	"context"
	"fmt"
	"time"

	"github.com/stretchr/testify/require"
)

func (s *ClientSuite) TestLogout_Success() {
	t := s.T()
	ctx := context.Background()

	var login = fmt.Sprintf("test-logout-%d@thingsdiary.io", time.Now().UnixMilli())

	// Arrange: Register and authenticate user
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	// Verify user is authenticated
	require.NotEmpty(t, s.client.authToken)

	// Act: Logout
	err = s.client.Logout(ctx)

	// Assert: Logout successful
	require.NoError(t, err)
	require.Empty(t, s.client.authToken)
	require.Nil(t, s.client.credentials)
}

func (s *ClientSuite) TestLogout_NotAuthenticated() {
	t := s.T()
	ctx := context.Background()

	// Act: Try to logout without authentication
	err := s.client.Logout(ctx)

	// Assert: Should fail with "not authenticated" error
	require.Error(t, err)
	require.Contains(t, err.Error(), "not authenticated")
}

func (s *ClientSuite) TestLogout_DoubleLogout() {
	t := s.T()
	ctx := context.Background()

	var login = fmt.Sprintf("test-double-logout-%d@thingsdiary.io", time.Now().UnixMilli())

	// Arrange: Register and authenticate user
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	// Act: First logout
	err = s.client.Logout(ctx)
	require.NoError(t, err)

	// Act: Second logout attempt
	err = s.client.Logout(ctx)

	// Assert: Second logout should fail
	require.Error(t, err)
	require.Contains(t, err.Error(), "not authenticated")
}

func (s *ClientSuite) TestLogout_ClearsCredentials() {
	t := s.T()
	ctx := context.Background()

	var login = fmt.Sprintf("test-logout-clear-%d@thingsdiary.io", time.Now().UnixMilli())

	// Arrange: Register and authenticate user
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)

	// Verify user has credentials and token
	require.NotEmpty(t, s.client.authToken)
	require.NotNil(t, s.client.credentials)

	// Act: Logout
	err = s.client.Logout(ctx)
	require.NoError(t, err)

	// Assert: Credentials and token are cleared
	require.Empty(t, s.client.authToken)
	require.Nil(t, s.client.credentials)
}
