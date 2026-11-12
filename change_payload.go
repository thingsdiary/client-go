package client

import (
	"github.com/pkg/errors"
	v1 "github.com/thingsdiary/client-go/proto"
	"google.golang.org/protobuf/proto"
)

// MarshalChangePayload encodes nonce and ciphertext as protobuf EncryptedData
// for Change.Payload. Encryption key metadata is sent separately.
func MarshalChangePayload(nonce, encryptedData []byte) ([]byte, error) {
	details := &v1.EncryptedData{
		Nonce: nonce,
		Data:  encryptedData,
	}
	if err := validateProtoEncryptedData(details, "change payload"); err != nil {
		return nil, errors.Wrap(err, "failed to marshal change payload")
	}

	payload, err := proto.Marshal(details)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal change payload")
	}

	return payload, nil
}

// UnmarshalChangePayload decodes a protobuf EncryptedData payload into nonce
// and ciphertext. It expects the inner Change.Payload, without the DAG wrapper.
func UnmarshalChangePayload(payload []byte) (nonce, encryptedData []byte, err error) {
	details := new(v1.EncryptedData)
	if err := proto.Unmarshal(payload, details); err != nil {
		return nil, nil, errors.Wrap(err, "failed to unmarshal change payload")
	}
	if err := validateProtoEncryptedData(details, "change payload"); err != nil {
		return nil, nil, errors.Wrap(err, "failed to unmarshal change payload")
	}

	return details.GetNonce(), details.GetData(), nil
}
