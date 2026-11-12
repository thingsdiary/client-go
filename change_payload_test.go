package client

import (
	"testing"

	"github.com/stretchr/testify/require"
	v1 "github.com/thingsdiary/client-go/proto"
	"google.golang.org/protobuf/proto"
)

func TestChangePayloadWireFormat(t *testing.T) {
	t.Parallel()
	nonce := []byte{0, 0xff}
	ciphertext := []byte{0x80, 0, 0xff}
	// Protobuf field 1 is nonce; field 2 is ciphertext. Both are length-delimited.
	fixture := []byte{0x0a, 2, 0, 0xff, 0x12, 3, 0x80, 0, 0xff}

	payload, err := MarshalChangePayload(nonce, ciphertext)
	require.NoError(t, err)
	require.Equal(t, fixture, payload)

	decodedNonce, decodedData, err := UnmarshalChangePayload(fixture)
	require.NoError(t, err)
	require.Equal(t, nonce, decodedNonce)
	require.Equal(t, ciphertext, decodedData)
}

func TestChangePayloadEncryptionRoundTrip(t *testing.T) {
	t.Parallel()
	key, err := generateSymmetricKey()
	require.NoError(t, err)
	plaintext, err := proto.Marshal(&v1.DiaryDetails{Title: "Private diary"})
	require.NoError(t, err)
	nonce, ciphertext, err := encryptWithSymmetricKey(plaintext, key)
	require.NoError(t, err)
	payload, err := MarshalChangePayload(nonce, ciphertext)
	require.NoError(t, err)
	decodedNonce, decodedData, err := UnmarshalChangePayload(payload)
	require.NoError(t, err)
	decrypted, err := decryptWithSymmetricKey(decodedNonce, decodedData, key)
	require.NoError(t, err)
	require.Equal(t, plaintext, decrypted)
}

func TestMarshalChangePayloadRejectsMissingFields(t *testing.T) {
	t.Parallel()
	for name, fields := range map[string]struct{ nonce, data []byte }{
		"missing nonce": {data: []byte{1}},
		"empty nonce":   {nonce: []byte{}, data: []byte{1}},
		"missing data":  {nonce: []byte{1}},
		"empty data":    {nonce: []byte{1}, data: []byte{}},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			payload, err := MarshalChangePayload(fields.nonce, fields.data)
			require.Error(t, err)
			require.Nil(t, payload)
		})
	}
}

func TestUnmarshalChangePayloadRejectsInvalidPayload(t *testing.T) {
	t.Parallel()
	for name, payload := range map[string][]byte{
		"missing payload": nil,
		"empty payload":   {},
		"malformed":       {0xff},
		"truncated":       {0x0a, 2, 1},
		"missing data":    {0x0a, 1, 1},
		"missing nonce":   {0x12, 1, 1},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			nonce, data, err := UnmarshalChangePayload(payload)
			require.Error(t, err)
			require.Nil(t, nonce)
			require.Nil(t, data)
		})
	}
}
