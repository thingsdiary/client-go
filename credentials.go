package client

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"io"

	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/hkdf"
	"golang.org/x/crypto/pbkdf2"
)

const (
	KdfSaltLength   = 32
	KdfIterations   = 650_000
	kdfSaltAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

	// Domain separation info for HKDF (RFC 5869).
	// Must match client-js `client-js/src/crypto/credentials.ts` exactly.
	hkdfInfoSigning    = "thingsdiary-ed25519-v1"
	hkdfInfoEncryption = "thingsdiary-x25519-v1"
)

type Credentials struct {
	EncryptionPublicKey  []byte
	EncryptionPrivateKey []byte

	SigningPublicKey  []byte
	SigningPrivateKey []byte
}

func NewCredentials(seedPhrase, salt string) (*Credentials, error) {
	// Step 1: PBKDF2-SHA256 to derive a master seed from the seed phrase.
	// This matches `client-js/src/crypto/credentials.ts`.
	masterSeed := pbkdf2.Key([]byte(seedPhrase), []byte(salt), KdfIterations, 32, sha256.New)

	// Step 2: HKDF-SHA256 domain separation.
	// We intentionally do not provide HKDF salt here (nil -> RFC 5869 default all-zero salt),
	// because PBKDF2 already applied the user-specific salt.
	signingSeed, err := hkdfExpand32(masterSeed, []byte(hkdfInfoSigning))
	if err != nil {
		return nil, err
	}

	encryptionSeed, err := hkdfExpand32(masterSeed, []byte(hkdfInfoEncryption))
	if err != nil {
		return nil, err
	}

	var privateKey [32]byte
	copy(privateKey[:], encryptionSeed[:32])
	var publicKey [32]byte

	curve25519.ScalarBaseMult(&publicKey, &privateKey)

	signPrivateKey := ed25519.NewKeyFromSeed(signingSeed)
	signPublicKey := signPrivateKey.Public().(ed25519.PublicKey)

	creds := Credentials{
		EncryptionPublicKey:  publicKey[:32],
		EncryptionPrivateKey: privateKey[:32],

		SigningPublicKey:  signPublicKey,
		SigningPrivateKey: signPrivateKey,
	}

	return &creds, nil
}

func hkdfExpand32(ikm, info []byte) ([]byte, error) {
	r := hkdf.New(sha256.New, ikm, nil, info)
	out := make([]byte, 32)

	if _, err := io.ReadFull(r, out); err != nil {
		return nil, err
	}

	return out, nil
}

func GenerateKdfSalt() (string, error) {
	bytes := make([]byte, KdfSaltLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	result := make([]byte, KdfSaltLength)
	for i, b := range bytes {
		result[i] = kdfSaltAlphabet[int(b)%len(kdfSaltAlphabet)]
	}

	return string(result), nil
}
