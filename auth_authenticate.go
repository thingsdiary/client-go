package client

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"

	"github.com/thingsdiary/client-go/openapi"
)

func (c *Client) Authenticate(ctx context.Context, login, password, seedPhrase string) error {
	loginResult, err := c.login(ctx, login, password)
	if err != nil {
		return errors.Wrap(err, "login failed")
	}

	credentials, err := NewCredentials(seedPhrase)
	if err != nil {
		return err
	}

	signedNonce := ed25519.Sign(credentials.SigningPrivateKey, loginResult.Nonce)
	verifyResult, err := c.loginVerify(ctx, loginResult.ChallengeId, signedNonce)
	if err != nil {
		return errors.Wrap(err, "login failed")
	}

	// Only set credentials and token after successful authentication
	c.credentials = credentials
	c.authToken = verifyResult.Token

	return nil
}

func (c *Client) login(ctx context.Context, login, password string) (*openapi.LoginResponse, error) {
	body := openapi.LoginRequest{
		Login:    login,
		Password: password,
	}

	url := fmt.Sprintf("%s/v1/auth/login", c.baseURL)
	req, err := c.newRequest(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrInvalidCredentials
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("response status code: %s", resp.Status)
	}

	var r openapi.LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}

	return &r, nil
}

func (c *Client) loginVerify(ctx context.Context, challengeID string, signedNonce []byte) (*openapi.LoginVerifyResponse, error) {
	body := openapi.LoginVerifyRequest{
		ChallengeId: challengeID,
		SignedNonce: signedNonce,
	}

	url := fmt.Sprintf("%s/v1/auth/login/verify", c.baseURL)
	req, err := c.newRequest(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		return nil, ErrInvalidChallenge
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("response status code: %s", resp.Status)
	}

	var r openapi.LoginVerifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}

	return &r, nil
}
