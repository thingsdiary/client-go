package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"

	"github.com/thingsdiary/client-go/openapi"
)

// GetRealtimeToken returns a short-lived JWT for a Centrifugo connection.
func (c *Client) GetRealtimeToken(ctx context.Context) (string, error) {
	req, err := c.newAuthenticatedRequest(ctx, http.MethodPost, fmt.Sprintf("%s/v1/realtime/token", c.baseURL), nil)
	if err != nil {
		return "", errors.Wrap(err, "failed to create realtime token request")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "failed to fetch realtime token")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return "", ErrUnauthorized
	}

	if resp.StatusCode != http.StatusOK {
		return "", errors.Errorf("unexpected realtime token status: %s", resp.Status)
	}

	var result openapi.RealtimeTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", errors.Wrap(err, "failed to decode realtime token")
	}

	if result.Token == "" {
		return "", errors.New("realtime token response is empty")
	}

	return result.Token, nil
}
