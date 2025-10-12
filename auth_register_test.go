package client

import (
	"context"
	"fmt"
	"time"

	"github.com/stretchr/testify/require"
)

func (s *ClientSuite) TestLogin_Register() {
	t := s.T()
	ctx := context.Background()

	var login = fmt.Sprintf("test-login-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
	require.NoError(t, err)
}
