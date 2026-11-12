package client

import (
	"context"
	"crypto/rand"
	"testing"

	"github.com/centrifugal/centrifuge-go"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	v1 "github.com/thingsdiary/client-go/proto"
	"golang.org/x/crypto/nacl/box"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestCreateDiary_EncryptsRootChange(t *testing.T) {
	// arrange
	publicKey, privateKey, err := box.GenerateKey(rand.Reader)
	require.NoError(t, err)

	c := NewClient()
	c.credentials = &Credentials{
		EncryptionPublicKey:  publicKey[:],
		EncryptionPrivateKey: privateKey[:],
		SigningPrivateKey:    testSigningPrivateKey(),
	}

	var method string
	var req v1.CreateDiaryRequest
	c.conn = &mockRealtimeConn{state: centrifuge.StateConnected, rpc: func(_ context.Context, gotMethod string, data []byte) (centrifuge.RPCResult, error) {
		var response proto.Message
		switch gotMethod {
		case "io.thingsdiary.v1.diary.create":
			method = gotMethod
			require.NoError(t, proto.Unmarshal(rpcPayload(t, gotMethod, data), &req))
			response = &v1.CreateDiaryResponse{DiaryId: "diary-1"}
		case "io.thingsdiary.v1.diary.get":
			response = &v1.GetDiaryResponse{Diary: &v1.EncryptedDiary{
				Id:             "diary-1",
				CreatedAt:      timestamppb.Now(),
				UpdatedAt:      timestamppb.Now(),
				EncryptionKeys: []*v1.DiaryEncryptionKey{{Id: "key-1", Value: req.EncryptedDiaryKey}},
				Change: &v1.EncryptedChange{
					Change:     req.GetEncryptedChange().GetChange(),
					Encryption: req.GetEncryptedChange().GetEncryption(),
				},
			}}
		default:
			t.Fatalf("unexpected RPC method: %s", gotMethod)
		}
		encoded, err := proto.Marshal(response)
		require.NoError(t, err)
		return centrifuge.RPCResult{Data: encoded}, nil
	}}

	// act
	diary, err := c.CreateDiary(context.Background(), CreateDiaryParams{
		Title:       "Private diary",
		Description: "Encrypted description",
	})

	// assert
	require.NoError(t, err)
	require.Equal(t, "io.thingsdiary.v1.diary.create", method)
	require.NotEmpty(t, req.GetEncryptedDiaryKey())
	require.NotEmpty(t, req.GetEncryptedChange().GetEncryption().GetEncryptedKeyNonce())
	require.NotEmpty(t, req.GetEncryptedChange().GetEncryption().GetEncryptedKeyData())
	require.NotNil(t, req.GetEncryptedChange().GetChange())
	_, err = uuid.Parse(req.GetEncryptedChange().GetChange().GetId())
	require.NoError(t, err)
	require.Equal(t, req.GetEncryptedChange().GetChange().GetId(), diary.ChangeID)
	require.Empty(t, req.GetEncryptedChange().GetChange().GetObjectId())
	require.Empty(t, req.GetEncryptedChange().GetChange().GetParentChangeIds())
	payloadNonce, payloadData, err := UnmarshalChangePayload(req.GetEncryptedChange().GetChange().GetPayload())
	require.NoError(t, err)
	require.NotEmpty(t, payloadNonce)
	require.NotEmpty(t, payloadData)

	diaryKey, ok := box.OpenAnonymous(nil, req.GetEncryptedDiaryKey(), publicKey, privateKey)
	require.True(t, ok)

	entityKey, err := decryptWithSymmetricKey(
		req.GetEncryptedChange().GetEncryption().GetEncryptedKeyNonce(),
		req.GetEncryptedChange().GetEncryption().GetEncryptedKeyData(),
		diaryKey,
	)
	require.NoError(t, err)

	detailsBytes, err := decryptWithSymmetricKey(payloadNonce, payloadData, entityKey)
	require.NoError(t, err)

	var details v1.DiaryDetails
	require.NoError(t, proto.Unmarshal(detailsBytes, &details))
	require.True(t, proto.Equal(&v1.DiaryDetails{
		Title:       "Private diary",
		Description: "Encrypted description",
	}, &details))
}
