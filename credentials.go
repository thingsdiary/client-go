package client

import (
	"crypto/ed25519"
	"crypto/sha256"

	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/pbkdf2"
)

type Credentials struct {
	EncryptionPublicKey  []byte
	EncryptionPrivateKey []byte

	SigningPublicKey  []byte
	SigningPrivateKey []byte
}

func NewCredentials(seedPhrase string) (*Credentials, error) {
	// todo: do not hardcode client secret
	salt := []byte("my-app-context")

	seed := pbkdf2.Key([]byte(seedPhrase), salt, 100_000, 32, sha256.New)

	var privateKey [32]byte
	copy(privateKey[:], seed[:32])
	var publicKey [32]byte

	curve25519.ScalarBaseMult(&publicKey, &privateKey)

	signPrivateKey := ed25519.NewKeyFromSeed(seed)
	signPublicKey := signPrivateKey.Public().(ed25519.PublicKey)

	creds := Credentials{
		EncryptionPublicKey:  publicKey[:32],
		EncryptionPrivateKey: privateKey[:32],

		SigningPublicKey:  signPublicKey,
		SigningPrivateKey: signPrivateKey,
	}

	return &creds, nil
}
