package client

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

type Client struct {
	baseURL     string
	credentials *Credentials
	httpClient  *http.Client
	authToken   string
}

func NewClient(opts ...clientOption) *Client {
	clientOptions := defaultOptions()
	for _, opt := range opts {
		opt(clientOptions)
	}

	client := Client{
		baseURL: clientOptions.baseURL,
		httpClient: &http.Client{
			Timeout: clientOptions.timeout,
		},
	}

	return &client
}

// newRequest creates an HTTP request with JSON body handling
func (c *Client) newRequest(ctx context.Context, method, url string, body interface{}) (*http.Request, error) {
	var bodyReader io.Reader = http.NoBody

	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, errors.Wrap(err, "failed to marshal request body")
		}

		bodyReader = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// newAuthenticatedRequest creates an HTTP request with authentication headers
func (c *Client) newAuthenticatedRequest(ctx context.Context, method, url string, body interface{}) (*http.Request, error) {
	if c.authToken == "" {
		return nil, errors.New("not authenticated")
	}

	req, err := c.newRequest(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.authToken)

	return req, nil
}
