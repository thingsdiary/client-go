package client

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ClientSuite struct {
	suite.Suite

	client *Client

	seedPhrase string
	kdfSalt    string
}

func TestClientSuite(t *testing.T) {
	suite.Run(t, new(ClientSuite))
}

func (s *ClientSuite) SetupSuite() {}

func (s *ClientSuite) SetupTest() {
	words := []string{
		"banana", "eagle", "mirror", "castle",
		"ocean", "rocket", "twist", "canyon",
		"zebra", "glue", "toast", "lemon",
	}

	s.seedPhrase = fmt.Sprint(words)

	kdfSalt, err := GenerateKdfSalt()
	if err != nil {
		s.T().Fatalf("failed to generate kdf salt: %v", err)
	}

	s.kdfSalt = kdfSalt

	baseURL := "http://localhost:8081/api"
	if envURL := os.Getenv("BASE_URL"); envURL != "" {
		baseURL = envURL
	}

	s.client = NewClient(WithBaseURL(baseURL))
}
