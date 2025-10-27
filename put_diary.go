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

// PutDiaryParams contains parameters for updating a diary
type PutDiaryParams struct {
	Title       string
	Description string
}

// GetDiaryDetails extracts diary details from parameters
func (p PutDiaryParams) GetDiaryDetails() DiaryDetails {
	return DiaryDetails(p)
}

func (c *Client) PutDiary(ctx context.Context, diaryID string, params PutDiaryParams) (*Diary, error) {
	if c.credentials == nil {
		return nil, ErrUnauthorized
	}

	diaryData, err := c.getDiary(ctx, diaryID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get current diary")
	}

	if len(diaryData.EncryptionKeys) == 0 {
		return nil, errors.New("no encryption keys found in diary")
	}

	diaryKeyID := diaryData.EncryptionKeys[0].Id

	entityKey, err := generateSymmetricKey()
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate entity key")
	}

	diaryDetails := params.GetDiaryDetails()
	diaryDetailsJSON, err := json.Marshal(diaryDetails)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal diary details")
	}

	contentNonce, encryptedContent, err := encryptWithSymmetricKey(diaryDetailsJSON, entityKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to encrypt diary content")
	}

	encryptedDiaryKeyValue := diaryData.EncryptionKeys[0].Value
	decryptedDiaryKey, err := decryptWithPrivateKey(
		encryptedDiaryKeyValue,
		c.credentials.EncryptionPrivateKey,
		c.credentials.EncryptionPublicKey,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decrypt diary key")
	}

	keyNonce, encryptedEntityKey, err := encryptWithSymmetricKey(entityKey, decryptedDiaryKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to encrypt entity key")
	}

	request := openapi.PutDiaryRequest{
		Version: NewVersion(),
		Details: openapi.EncryptedData{
			Nonce: contentNonce,
			Data:  encryptedContent,
		},
		Encryption: openapi.DiaryEncryption{
			DiaryKeyId:        diaryKeyID,
			EncryptedKeyNonce: keyNonce,
			EncryptedKeyData:  encryptedEntityKey,
		},
	}

	if err := request.Validate(); err != nil {
		return nil, errors.Wrap(err, "request validation failed")
	}

	url := fmt.Sprintf("%s/v1/diaries/%s", c.baseURL, diaryID)
	req, err := c.newAuthenticatedRequest(ctx, http.MethodPut, url, request)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}

	requestJSON, err := json.Marshal(request)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal request")
	}

	signature := signBytes(requestJSON, c.credentials.SigningPrivateKey)
	req.Header.Set("X-Signature", base64.StdEncoding.EncodeToString(signature))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute request")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrDiaryNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("unexpected status code: %s", resp.Status)
	}

	var apiResponse openapi.PutDiaryResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, errors.Wrap(err, "failed to decode response")
	}

	if len(apiResponse.Diary.EncryptionKeys) == 0 {
		return nil, errors.New("no encryption keys found in response")
	}

	encryptedDiaryKeyValue = apiResponse.Diary.EncryptionKeys[0].Value
	decryptedDiaryKey, err = decryptWithPrivateKey(
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
