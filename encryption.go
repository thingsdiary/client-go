package client

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"

	"github.com/pkg/errors"
	"golang.org/x/crypto/nacl/box"
)

// generateSymmetricKey generates a 32-byte AES-256 key
func generateSymmetricKey() ([]byte, error) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate symmetric key")
	}
	return key, nil
}

// encryptWithSymmetricKey encrypts data using AES-256-GCM
func encryptWithSymmetricKey(data []byte, key []byte) (nonce []byte, ciphertext []byte, err error) {
	if len(key) != 32 {
		return nil, nil, errors.New("key must be 32 bytes for AES-256")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to create AES cipher")
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to create GCM")
	}

	nonce = make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, errors.Wrap(err, "failed to generate nonce")
	}

	ciphertext = gcm.Seal(nil, nonce, data, nil)
	return nonce, ciphertext, nil
}

// decryptWithSymmetricKey decrypts data using AES-256-GCM
func decryptWithSymmetricKey(nonce []byte, ciphertext []byte, key []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, errors.New("key must be 32 bytes for AES-256")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create AES cipher")
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create GCM")
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decrypt data")
	}

	return plaintext, nil
}

// encryptWithPublicKey encrypts data using NaCl box.SealAnonymous for envelope encryption
func encryptWithPublicKey(data []byte, publicKey []byte) ([]byte, error) {
	if len(publicKey) != 32 {
		return nil, errors.New("public key must be 32 bytes")
	}

	var pubKey [32]byte
	copy(pubKey[:], publicKey)

	encrypted, err := box.SealAnonymous(nil, data, &pubKey, rand.Reader)
	if err != nil {
		return nil, errors.Wrap(err, "failed to encrypt with public key")
	}

	return encrypted, nil
}

// decryptWithPrivateKey decrypts data using NaCl box.OpenAnonymous
func decryptWithPrivateKey(encrypted []byte, privateKey []byte, publicKey []byte) ([]byte, error) {
	if len(privateKey) != 32 {
		return nil, errors.New("private key must be 32 bytes")
	}
	if len(publicKey) != 32 {
		return nil, errors.New("public key must be 32 bytes")
	}

	var privKey [32]byte
	var pubKey [32]byte
	copy(privKey[:], privateKey)
	copy(pubKey[:], publicKey)

	decrypted, ok := box.OpenAnonymous(nil, encrypted, &pubKey, &privKey)
	if !ok {
		return nil, errors.New("failed to decrypt with private key")
	}

	return decrypted, nil
}
