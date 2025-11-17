package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"

	"github.com/thingsdiary/client-go/openapi"
)

func (c *Client) Register(ctx context.Context, login, password, seedPhrase string) error {
	creds, err := NewCredentials(seedPhrase)
	if err != nil {
		return err
	}

	body := openapi.RegisterRequest{
		Login:               login,
		Password:            password,
		SignaturePublicKey:  creds.SigningPublicKey,
		EncryptionPublicKey: creds.EncryptionPublicKey,
	}

	url := fmt.Sprintf("%s/v1/auth/register", c.baseURL)
	req, err := c.newRequest(ctx, http.MethodPost, url, body)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		return ErrAccountAlreadyExists
	}

	if resp.StatusCode == http.StatusBadRequest {
		var errResp openapi.ErrorResponse

		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return errors.Errorf("register failed: %s", resp.Status)
		}

		if errResp.ErrorCode == openapi.ResponseErrorCodeInvalidLogin {
			return ErrInvalidLogin
		}

		if errResp.ErrorCode == openapi.ResponseErrorCodeInvalidPassword {
			return ErrInvalidPassword
		}

		return errors.Errorf("register failed: %s", resp.Status)
	}

	if resp.StatusCode != http.StatusCreated {
		return errors.Errorf("register failed: %s", resp.Status)
	}

	var r openapi.RegisterResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return err
	}

	return nil
}
