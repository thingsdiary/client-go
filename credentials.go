package client

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"

	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/pbkdf2"
)

const (
	KdfSaltLength   = 32
	KdfIterations   = 650_000
	kdfSaltAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
)

type Credentials struct {
	EncryptionPublicKey  []byte
	EncryptionPrivateKey []byte

	SigningPublicKey  []byte
	SigningPrivateKey []byte
}

func NewCredentials(seedPhrase, salt string) (*Credentials, error) {
	seed := pbkdf2.Key([]byte(seedPhrase), []byte(salt), KdfIterations, 32, sha256.New)

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
