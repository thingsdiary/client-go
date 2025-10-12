package client

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/thingsdiary/client-go/openapi"
)

func (s *ClientSuite) TestCreateDiary_SignatureVerification() {
	t := s.T()
	ctx := context.Background()

	t.Run("valid_signature_success", func(t *testing.T) {
		// Arrange: Register and authenticate user
		var login = fmt.Sprintf("test-create-diary-sig-valid-%d@thingsdiary.io", time.Now().UnixMilli())
		err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
		require.NoError(t, err)

		err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
		require.NoError(t, err)

		// Act: Create diary with valid signature
		diary, err := s.client.CreateDiary(ctx, CreateDiaryParams{
			Title:       "Test Diary",
			Description: "Diary for signature testing",
		})

		// Assert: Diary creation successful
		require.NoError(t, err)
		assert.Equal(t, "Test Diary", diary.Title)
		assert.Equal(t, "Diary for signature testing", diary.Description)
		assert.NotEmpty(t, diary.ID)
	})

	t.Run("missing_signature_header", func(t *testing.T) {
		// Arrange: Register and authenticate user
		var login = fmt.Sprintf("test-create-diary-sig-missing-%d@thingsdiary.io", time.Now().UnixMilli())
		err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
		require.NoError(t, err)

		err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
		require.NoError(t, err)

		// Act: Create request without X-Signature header
		request := openapi.CreateDiaryRequest{
			EncryptedDiaryKey: []byte("test-encrypted-key"),
			Details: openapi.EncryptedData{
				Nonce: []byte("test-nonce"),
				Data:  []byte("test-data"),
			},
		}

		url := fmt.Sprintf("%s/api/v1/diaries", s.client.baseURL)
		req, err := s.client.newAuthenticatedRequest(ctx, http.MethodPost, url, request)
		require.NoError(t, err)

		// Remove X-Signature header to test missing signature
		req.Header.Del("X-Signature")

		// Act: Send request without signature
		resp, err := s.client.httpClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Assert: Should return 400 with INVALID_SIGNATURE error
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var errorResp openapi.ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&errorResp)
		require.NoError(t, err)
		assert.Equal(t, openapi.ResponseErrorCodeInvalidSignature, errorResp.ErrorCode)
	})

	t.Run("invalid_base64_signature", func(t *testing.T) {
		// Arrange: Register and authenticate user
		var login = fmt.Sprintf("test-create-diary-sig-invalid-b64-%d@thingsdiary.io", time.Now().UnixMilli())
		err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
		require.NoError(t, err)

		err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
		require.NoError(t, err)

		// Act: Create request with invalid base64 signature
		request := openapi.CreateDiaryRequest{
			EncryptedDiaryKey: []byte("test-encrypted-key"),
			Details: openapi.EncryptedData{
				Nonce: []byte("test-nonce"),
				Data:  []byte("test-data"),
			},
		}

		url := fmt.Sprintf("%s/api/v1/diaries", s.client.baseURL)
		req, err := s.client.newAuthenticatedRequest(ctx, http.MethodPost, url, request)
		require.NoError(t, err)

		// Set invalid base64 signature
		req.Header.Set("X-Signature", "invalid-base64!")

		// Act: Send request with invalid signature
		resp, err := s.client.httpClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Assert: Should return 400 with INVALID_SIGNATURE error
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var errorResp openapi.ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&errorResp)
		require.NoError(t, err)
		assert.Equal(t, openapi.ResponseErrorCodeInvalidSignature, errorResp.ErrorCode)
	})

	t.Run("invalid_signature_content", func(t *testing.T) {
		// Arrange: Register and authenticate user
		var login = fmt.Sprintf("test-create-diary-sig-invalid-%d@thingsdiary.io", time.Now().UnixMilli())
		err := s.client.Register(ctx, login, "password-123", s.seedPhrase)
		require.NoError(t, err)

		err = s.client.Authenticate(ctx, login, "password-123", s.seedPhrase)
		require.NoError(t, err)

		// Act: Create request with invalid signature content
		request := openapi.CreateDiaryRequest{
			EncryptedDiaryKey: []byte("test-encrypted-key"),
			Details: openapi.EncryptedData{
				Nonce: []byte("test-nonce"),
				Data:  []byte("test-data"),
			},
		}

		url := fmt.Sprintf("%s/api/v1/diaries", s.client.baseURL)
		req, err := s.client.newAuthenticatedRequest(ctx, http.MethodPost, url, request)
		require.NoError(t, err)

		// Set invalid signature (valid base64 but wrong signature)
		invalidSignature := make([]byte, 64) // Ed25519 signature size
		req.Header.Set("X-Signature", base64.StdEncoding.EncodeToString(invalidSignature))

		// Act: Send request with invalid signature
		resp, err := s.client.httpClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Assert: Should return 400 with INVALID_SIGNATURE error
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var errorResp openapi.ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&errorResp)
		require.NoError(t, err)
		assert.Equal(t, openapi.ResponseErrorCodeInvalidSignature, errorResp.ErrorCode)
	})
}
