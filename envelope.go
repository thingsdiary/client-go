package client

import (
	"crypto/ed25519"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"

	v1 "github.com/thingsdiary/client-go/proto"
)

const envelopeVersion uint32 = 1

// NewEnvelope binds an encoded RPC payload to its method and logical operation.
func NewEnvelope(method string, payload []byte) *v1.Envelope {
	return &v1.Envelope{
		Version:     envelopeVersion,
		Method:      method,
		Payload:     payload,
		OperationId: uuid.NewString(),
	}
}

func signEnvelope(envelope *v1.Envelope, privateKey ed25519.PrivateKey) (*v1.SignedEnvelope, error) {
	if len(privateKey) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid Ed25519 signing private key")
	}

	data, err := proto.Marshal(envelope)
	if err != nil {
		return nil, fmt.Errorf("encode RPC envelope: %w", err)
	}

	return &v1.SignedEnvelope{
		Envelope:  data,
		Signature: signBytes(data, privateKey),
	}, nil
}
