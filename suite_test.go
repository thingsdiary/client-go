package client

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ClientSuite struct {
	suite.Suite

	client *Client

	seedPhrase string
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

	s.client = NewClient(WithBaseURL("http://localhost:8080"))
}
