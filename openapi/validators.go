package openapi

import (
	"crypto/ed25519"
	"errors"
	"fmt"
)

// Version validation constants
const (
	// MinValidVersion represents the minimum valid version (Jan 1, 2020 UTC)
	MinValidVersion = uint64(1577836800000) // 2020-01-01T00:00:00Z

	// MaxValidVersion represents the maximum valid version (Jan 1, 2100 UTC)
	MaxValidVersion = uint64(4102444800000) // 2100-01-01T00:00:00Z
)

// validateVersion validates that a version number is within reasonable bounds
func validateVersion(version uint64) error {
	if version < MinValidVersion {
		return fmt.Errorf("version %d is too low, must be >= %d (2020-01-01)", version, MinValidVersion)
	}

	if version > MaxValidVersion {
		return fmt.Errorf("version %d is too high, must be <= %d (2100-01-01)", version, MaxValidVersion)
	}

	return nil
}

func (r *RegisterRequest) Validate() error {
	// All validation happens on the server:
	// - Login format (email)
	// - Password strength (min 8 chars)
	// - Encryption key (Curve25519 public key)
	// - Signature key (Ed25519 public key)
	return nil
}

func (r *LoginRequest) Validate() error {
	// All validation happens on the server
	// Server returns only INVALID_CREDENTIALS for security (no details exposed)
	return nil
}

func (r *LoginVerifyRequest) Validate() error {
	if r.ChallengeId == "" {
		return errors.New("challenge_id is required")
	}

	if len(r.SignedNonce) != ed25519.SignatureSize {
		return fmt.Errorf("invalid signed_nonce length: expected %d, got %d", ed25519.SignatureSize, len(r.SignedNonce))
	}

	return nil
}

func (r *CreateDiaryRequest) Validate() error {
	if len(r.EncryptedDiaryKey) == 0 {
		return errors.New("encrypted_diary_key is required")
	}

	if len(r.Details.Nonce) == 0 {
		return errors.New("details_nonce is required")
	}

	if len(r.Details.Data) == 0 {
		return errors.New("details_data is required")
	}

	return nil
}

func (e *DiaryEncryption) Validate() error {
	if e.DiaryKeyId == "" {
		return errors.New("diary key id is required")
	}

	if len(e.EncryptedKeyNonce) == 0 {
		return errors.New("encrypted key nonce is required")
	}

	if len(e.EncryptedKeyData) == 0 {
		return errors.New("encrypted key data is required")
	}

	return nil
}

func (r *PutDiaryRequest) Validate() error {
	if err := validateVersion(r.Version); err != nil {
		return err
	}

	if err := r.Encryption.Validate(); err != nil {
		return err
	}

	if len(r.Details.Nonce) == 0 {
		return errors.New("details.nonce is required")
	}

	if len(r.Details.Data) == 0 {
		return errors.New("details.data is required")
	}

	return nil
}

func (r *PutEntryRequest) Validate() error {
	if err := validateVersion(r.Version); err != nil {
		return err
	}

	if err := r.Encryption.Validate(); err != nil {
		return err
	}

	if len(r.Details.Nonce) == 0 {
		return errors.New("details.nonce is required")
	}

	if len(r.Details.Data) == 0 {
		return errors.New("details.data is required")
	}

	if len(r.Preview.Nonce) == 0 {
		return errors.New("preview.nonce is required")
	}

	if len(r.Preview.Data) == 0 {
		return errors.New("preview.data is required")
	}

	return nil
}

func (r *PutTopicRequest) Validate() error {
	if err := validateVersion(r.Version); err != nil {
		return err
	}

	if err := r.Encryption.Validate(); err != nil {
		return err
	}

	if len(r.Details.Nonce) == 0 {
		return errors.New("details.nonce is required")
	}

	if len(r.Details.Data) == 0 {
		return errors.New("details.data is required")
	}

	return nil
}

func (r *PutTemplateRequest) Validate() error {
	if err := validateVersion(r.Version); err != nil {
		return err
	}

	if err := r.Encryption.Validate(); err != nil {
		return err
	}

	if len(r.Details.Nonce) == 0 {
		return errors.New("details.nonce is required")
	}

	if len(r.Details.Data) == 0 {
		return errors.New("details.data is required")
	}

	return nil
}
