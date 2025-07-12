package client

import (
	"crypto/ed25519"
)

// signBytes signs data with Ed25519 private key
func signBytes(data []byte, privateKey ed25519.PrivateKey) []byte {
	return ed25519.Sign(privateKey, data)
}
