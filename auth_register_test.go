package client

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRegister_BadRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	kdfSalt, err := GenerateKdfSalt()
	require.NoError(t, err)

	client := NewClient(WithBaseURL(server.URL))
	err = client.Register(context.Background(), "promo.koko@gmail.com", "qwerty", "seed-phrase", kdfSalt)

	require.ErrorIs(t, err, ErrBadRequest)
}

func (s *ClientSuite) TestLogin_Register() {
	t := s.T()
	ctx := context.Background()

	var login = fmt.Sprintf("test-login-%d@thingsdiary.io", time.Now().UnixMilli())
	err := s.client.Register(ctx, login, "password-123", s.seedPhrase, s.kdfSalt)
	require.NoError(t, err)
}
