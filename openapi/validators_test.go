package openapi

import (
	"crypto/ed25519"
	"crypto/sha256"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/pbkdf2"
)

func TestValidateVersion(t *testing.T) {
	testCases := []struct {
		name        string
		version     uint64
		expectError bool
	}{
		{
			name:        "valid current timestamp",
			version:     uint64(time.Now().UnixMilli()),
			expectError: false,
		},
		{
			name:        "valid minimum version",
			version:     MinValidVersion,
			expectError: false,
		},
		{
			name:        "valid maximum version",
			version:     MaxValidVersion,
			expectError: false,
		},
		{
			name:        "version too low",
			version:     MinValidVersion - 1,
			expectError: true,
		},
		{
			name:        "version too high",
			version:     MaxValidVersion + 1,
			expectError: true,
		},
		{
			name:        "zero version",
			version:     0,
			expectError: true,
		},
		{
			name:        "extremely high version",
			version:     ^uint64(0), // uint64 max
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateVersion(tc.version)

			if tc.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "version")
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestPutDiaryRequest_Validate_Version(t *testing.T) {
	validRequest := &PutDiaryRequest{
		Version: uint64(time.Now().UnixMilli()),
		Encryption: DiaryEncryption{
			DiaryKeyId:        "test-key-id",
			EncryptedKeyNonce: []byte("nonce"),
			EncryptedKeyData:  []byte("data"),
		},
		Details: EncryptedData{
			Nonce: []byte("details-nonce"),
			Data:  []byte("details-data"),
		},
	}

	testCases := []struct {
		name        string
		version     uint64
		expectError bool
	}{
		{
			name:        "valid version",
			version:     uint64(time.Now().UnixMilli()),
			expectError: false,
		},
		{
			name:        "version too low",
			version:     MinValidVersion - 1,
			expectError: true,
		},
		{
			name:        "version too high",
			version:     MaxValidVersion + 1,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := *validRequest
			req.Version = tc.version

			err := req.Validate()

			if tc.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "version")
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestPutEntryRequest_Validate_Version(t *testing.T) {
	validRequest := &PutEntryRequest{
		Version: uint64(time.Now().UnixMilli()),
		Encryption: DiaryEncryption{
			DiaryKeyId:        "test-key-id",
			EncryptedKeyNonce: []byte("nonce"),
			EncryptedKeyData:  []byte("data"),
		},
		Details: EncryptedData{
			Nonce: []byte("details-nonce"),
			Data:  []byte("details-data"),
		},
		Preview: EncryptedData{
			Nonce: []byte("preview-nonce"),
			Data:  []byte("preview-data"),
		},
	}

	testCases := []struct {
		name        string
		version     uint64
		expectError bool
	}{
		{
			name:        "valid version",
			version:     uint64(time.Now().UnixMilli()),
			expectError: false,
		},
		{
			name:        "version too low",
			version:     MinValidVersion - 1,
			expectError: true,
		},
		{
			name:        "version too high",
			version:     MaxValidVersion + 1,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := *validRequest
			req.Version = tc.version

			err := req.Validate()

			if tc.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "version")
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestPutTopicRequest_Validate_Version(t *testing.T) {
	validRequest := &PutTopicRequest{
		Version: uint64(time.Now().UnixMilli()),
		Encryption: DiaryEncryption{
			DiaryKeyId:        "test-key-id",
			EncryptedKeyNonce: []byte("nonce"),
			EncryptedKeyData:  []byte("data"),
		},
		Details: EncryptedData{
			Nonce: []byte("details-nonce"),
			Data:  []byte("details-data"),
		},
	}

	testCases := []struct {
		name        string
		version     uint64
		expectError bool
	}{
		{
			name:        "valid version",
			version:     uint64(time.Now().UnixMilli()),
			expectError: false,
		},
		{
			name:        "version too low",
			version:     MinValidVersion - 1,
			expectError: true,
		},
		{
			name:        "version too high",
			version:     MaxValidVersion + 1,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := *validRequest
			req.Version = tc.version

			err := req.Validate()

			if tc.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "version")
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestPutTemplateRequest_Validate_Version(t *testing.T) {
	validRequest := &PutTemplateRequest{
		Version: uint64(time.Now().UnixMilli()),
		Encryption: DiaryEncryption{
			DiaryKeyId:        "test-key-id",
			EncryptedKeyNonce: []byte("nonce"),
			EncryptedKeyData:  []byte("data"),
		},
		Details: EncryptedData{
			Nonce: []byte("details-nonce"),
			Data:  []byte("details-data"),
		},
	}

	testCases := []struct {
		name        string
		version     uint64
		expectError bool
	}{
		{
			name:        "valid version",
			version:     uint64(time.Now().UnixMilli()),
			expectError: false,
		},
		{
			name:        "version too low",
			version:     MinValidVersion - 1,
			expectError: true,
		},
		{
			name:        "version too high",
			version:     MaxValidVersion + 1,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := *validRequest
			req.Version = tc.version

			err := req.Validate()

			if tc.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "version")
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestLoginVerifyRequest_Validate(t *testing.T) {
	validSignature := make([]byte, ed25519.SignatureSize)
	copy(validSignature, []byte("valid_signature_64_bytes_long_for_ed25519_signature_validation"))

	testCases := []struct {
		name        string
		challengeId string
		signedNonce []byte
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid request",
			challengeId: "test-challenge-id",
			signedNonce: validSignature,
			expectError: false,
		},
		{
			name:        "empty challenge_id",
			challengeId: "",
			signedNonce: validSignature,
			expectError: true,
			errorMsg:    "challenge_id is required",
		},
		{
			name:        "invalid signed_nonce length - too short",
			challengeId: "test-challenge-id",
			signedNonce: []byte("short"),
			expectError: true,
			errorMsg:    "invalid signed_nonce length",
		},
		{
			name:        "invalid signed_nonce length - too long",
			challengeId: "test-challenge-id",
			signedNonce: make([]byte, ed25519.SignatureSize+1),
			expectError: true,
			errorMsg:    "invalid signed_nonce length",
		},
		{
			name:        "empty signed_nonce",
			challengeId: "test-challenge-id",
			signedNonce: []byte{},
			expectError: true,
			errorMsg:    "invalid signed_nonce length",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := &LoginVerifyRequest{
				ChallengeId: tc.challengeId,
				SignedNonce: tc.signedNonce,
			}

			err := req.Validate()

			if tc.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRegisterRequest_Validate(t *testing.T) {
	// Generate valid keys
	seedPhrase := "test-seed-phrase-for-testing"
	salt := []byte("my-app-context")
	seed := pbkdf2.Key([]byte(seedPhrase), salt, 100_000, 32, sha256.New)

	// Valid encryption key (Curve25519)
	var encPrivateKey [32]byte
	copy(encPrivateKey[:], seed[:32])
	var encPublicKey [32]byte
	curve25519.ScalarBaseMult(&encPublicKey, &encPrivateKey)

	// Valid signing key (Ed25519)
	signPrivateKey := ed25519.NewKeyFromSeed(seed)
	signPublicKey := signPrivateKey.Public().(ed25519.PublicKey)

	testCases := []struct {
		name             string
		login            string
		password         string
		encryptionPubKey []byte
		signaturePubKey  []byte
		expectError      bool
		errorMsg         string
	}{
		{
			name:             "valid request",
			login:            "testuser",
			password:         "testpassword",
			encryptionPubKey: encPublicKey[:],
			signaturePubKey:  signPublicKey,
			expectError:      false,
		},
		{
			name:             "empty login",
			login:            "",
			password:         "testpassword",
			encryptionPubKey: encPublicKey[:],
			signaturePubKey:  signPublicKey,
			expectError:      true,
			errorMsg:         "login must not be empty",
		},
		{
			name:             "whitespace only login",
			login:            "   ",
			password:         "testpassword",
			encryptionPubKey: encPublicKey[:],
			signaturePubKey:  signPublicKey,
			expectError:      true,
			errorMsg:         "login must not be empty",
		},
		{
			name:             "empty password",
			login:            "testuser",
			password:         "",
			encryptionPubKey: encPublicKey[:],
			signaturePubKey:  signPublicKey,
			expectError:      true,
			errorMsg:         "password must not be empty",
		},
		{
			name:             "whitespace only password",
			login:            "testuser",
			password:         "   ",
			encryptionPubKey: encPublicKey[:],
			signaturePubKey:  signPublicKey,
			expectError:      true,
			errorMsg:         "password must not be empty",
		},
		{
			name:             "invalid encryption key - wrong length",
			login:            "testuser",
			password:         "testpassword",
			encryptionPubKey: []byte("short"),
			signaturePubKey:  signPublicKey,
			expectError:      true,
			errorMsg:         "invalid Curve25519 encryption public key",
		},
		{
			name:             "invalid encryption key - all zeros",
			login:            "testuser",
			password:         "testpassword",
			encryptionPubKey: make([]byte, 32),
			signaturePubKey:  signPublicKey,
			expectError:      true,
			errorMsg:         "invalid Curve25519 encryption public key",
		},
		{
			name:             "invalid signature key - wrong length",
			login:            "testuser",
			password:         "testpassword",
			encryptionPubKey: encPublicKey[:],
			signaturePubKey:  []byte("short"),
			expectError:      true,
			errorMsg:         "invalid Ed25519 signature public key",
		},
		{
			name:             "invalid signature key - random bytes",
			login:            "testuser",
			password:         "testpassword",
			encryptionPubKey: encPublicKey[:],
			signaturePubKey:  make([]byte, 32),
			expectError:      true,
			errorMsg:         "invalid Ed25519 signature public key",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := &RegisterRequest{
				Login:               tc.login,
				Password:            tc.password,
				EncryptionPublicKey: tc.encryptionPubKey,
				SignaturePublicKey:  tc.signaturePubKey,
			}

			err := req.Validate()

			if tc.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestIsValidEncryptionPublicKey(t *testing.T) {
	// Generate valid key
	seedPhrase := "test-seed-phrase"
	salt := []byte("my-app-context")
	seed := pbkdf2.Key([]byte(seedPhrase), salt, 100_000, 32, sha256.New)

	var privateKey [32]byte
	copy(privateKey[:], seed[:32])
	var publicKey [32]byte
	curve25519.ScalarBaseMult(&publicKey, &privateKey)

	testCases := []struct {
		name  string
		key   []byte
		valid bool
	}{
		{
			name:  "valid key",
			key:   publicKey[:],
			valid: true,
		},
		{
			name:  "invalid - too short",
			key:   []byte("short"),
			valid: false,
		},
		{
			name:  "invalid - too long",
			key:   make([]byte, 33),
			valid: false,
		},
		{
			name:  "invalid - all zeros",
			key:   make([]byte, 32),
			valid: false,
		},
		{
			name:  "invalid - empty",
			key:   []byte{},
			valid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isValidEncryptionPublicKey(tc.key)
			assert.Equal(t, tc.valid, result)
		})
	}
}

func TestIsValidSignaturePublicKey(t *testing.T) {
	// Generate valid Ed25519 key
	seedPhrase := "test-seed-phrase"
	salt := []byte("my-app-context")
	seed := pbkdf2.Key([]byte(seedPhrase), salt, 100_000, 32, sha256.New)

	privateKey := ed25519.NewKeyFromSeed(seed)
	publicKey := privateKey.Public().(ed25519.PublicKey)

	testCases := []struct {
		name  string
		key   []byte
		valid bool
	}{
		{
			name:  "valid key",
			key:   publicKey,
			valid: true,
		},
		{
			name:  "invalid - too short",
			key:   []byte("short"),
			valid: false,
		},
		{
			name:  "invalid - too long",
			key:   make([]byte, 33),
			valid: false,
		},
		{
			name:  "invalid - all zeros (not a valid point)",
			key:   make([]byte, 32),
			valid: false,
		},
		{
			name:  "invalid - empty",
			key:   []byte{},
			valid: false,
		},
		{
			name:  "invalid - random bytes (not a valid point)",
			key:   []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
			valid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isValidSignaturePublicKey(tc.key)
			assert.Equal(t, tc.valid, result)
		})
	}
}
