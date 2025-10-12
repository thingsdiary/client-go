package client

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"

	"github.com/thingsdiary/client-go/openapi"
)

// CreateDiaryParams contains parameters for creating a new diary
type CreateDiaryParams struct {
	Title       string
	Description string
}

// GetDiaryDetails extracts diary details from parameters
func (p CreateDiaryParams) GetDiaryDetails() DiaryDetails {
	return DiaryDetails(p)
}

// CreateDiary creates a new diary with zero-knowledge encryption
func (c *Client) CreateDiary(ctx context.Context, params CreateDiaryParams) (*Diary, error) {
	if c.credentials == nil {
		return nil, ErrUnauthorized
	}

	// Generate diary key for envelope encryption
	diaryKey, err := generateSymmetricKey()
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate diary key")
	}

	// Generate entity key for content encryption
	entityKey, err := generateSymmetricKey()
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate entity key")
	}

	// Encrypt diary content with entity key
	diaryDetails := params.GetDiaryDetails()
	diaryDetailsJSON, err := json.Marshal(diaryDetails)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal diary details")
	}

	contentNonce, encryptedContent, err := encryptWithSymmetricKey(diaryDetailsJSON, entityKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to encrypt diary content")
	}

	// Encrypt entity key with diary key
	keyNonce, encryptedEntityKey, err := encryptWithSymmetricKey(entityKey, diaryKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to encrypt entity key")
	}

	// Encrypt diary key with user's public key (envelope encryption)
	encryptedDiaryKey, err := encryptWithPublicKey(diaryKey, c.credentials.EncryptionPublicKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to encrypt diary key")
	}

	request := openapi.CreateDiaryRequest{
		EncryptedDiaryKey: encryptedDiaryKey,
		Details: openapi.EncryptedData{
			Nonce: contentNonce,
			Data:  encryptedContent,
		},
		Encryption: struct {
			EncryptedKeyData  []byte `json:"encrypted_key_data"`
			EncryptedKeyNonce []byte `json:"encrypted_key_nonce"`
		}{
			EncryptedKeyData:  encryptedEntityKey,
			EncryptedKeyNonce: keyNonce,
		},
	}

	requestJSON, err := json.Marshal(request)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal request")
	}

	url := fmt.Sprintf("%s/api/v1/diaries", c.baseURL)
	req, err := c.newAuthenticatedRequest(ctx, http.MethodPost, url, request)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}

	signature := signBytes(requestJSON, c.credentials.SigningPrivateKey)
	req.Header.Set("X-Signature", base64.StdEncoding.EncodeToString(signature))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute request")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadRequest {
		var errorResp openapi.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil {
			if errorResp.ErrorCode == openapi.ResponseErrorCodeDiaryLimitExceeded {
				return nil, ErrDiaryLimitExceeded
			}
		}

		return nil, errors.Errorf("bad request: %s", resp.Status)
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, errors.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var apiResponse openapi.CreateDiaryResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, errors.Wrap(err, "failed to decode response")
	}

	if len(apiResponse.Diary.EncryptionKeys) == 0 {
		return nil, errors.New("no encryption keys found in response")
	}

	encryptedDiaryKeyValue := apiResponse.Diary.EncryptionKeys[0].Value
	decryptedDiaryKey, err := decryptWithPrivateKey(
		encryptedDiaryKeyValue,
		c.credentials.EncryptionPrivateKey,
		c.credentials.EncryptionPublicKey,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decrypt diary key")
	}

	decryptedEntityKey, err := decryptWithSymmetricKey(
		apiResponse.Diary.Encryption.EncryptedKeyNonce,
		apiResponse.Diary.Encryption.EncryptedKeyData,
		decryptedDiaryKey,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decrypt entity key")
	}

	decryptedContentBytes, err := decryptWithSymmetricKey(
		apiResponse.Diary.Details.Nonce,
		apiResponse.Diary.Details.Data,
		decryptedEntityKey,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decrypt diary content")
	}

	var decryptedDiaryDetails DiaryDetails
	if err := json.Unmarshal(decryptedContentBytes, &decryptedDiaryDetails); err != nil {
		return nil, errors.Wrap(err, "failed to parse decrypted diary details")
	}

	diary := Diary{
		ID:          apiResponse.Diary.Id,
		Title:       decryptedDiaryDetails.Title,
		Description: decryptedDiaryDetails.Description,
		CreatedAt:   apiResponse.Diary.CreatedAt,
		UpdatedAt:   apiResponse.Diary.UpdatedAt,
		Version:     apiResponse.Diary.Version,
	}

	return &diary, nil
}
