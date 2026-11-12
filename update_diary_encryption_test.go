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

func TestUpdateDiary_EncryptsChildChange(t *testing.T) {
	// arrange
	publicKey, privateKey, err := box.GenerateKey(rand.Reader)
	require.NoError(t, err)
	diaryKey, err := generateSymmetricKey()
	require.NoError(t, err)
	envelope, err := encryptWithPublicKey(diaryKey, publicKey[:])
	require.NoError(t, err)

	c := NewClient()
	c.credentials = &Credentials{
		EncryptionPublicKey:  publicKey[:],
		EncryptionPrivateKey: privateKey[:],
		SigningPrivateKey:    testSigningPrivateKey(),
	}
	parents := []string{"head-a", "head-b"}
	var methods []string
	var req v1.PutDiaryRequest
	putReceived := false
	c.conn = &mockRealtimeConn{state: centrifuge.StateConnected, rpc: func(_ context.Context, method string, data []byte) (centrifuge.RPCResult, error) {
		methods = append(methods, method)
		var response proto.Message
		switch method {
		case "io.thingsdiary.v1.diary.get":
			var getReq v1.GetDiaryRequest
			require.NoError(t, proto.Unmarshal(rpcPayload(t, method, data), &getReq))
			require.Equal(t, "diary-1", getReq.GetDiaryId())
			response = &v1.GetDiaryResponse{Diary: &v1.EncryptedDiary{
				Id:             "diary-1",
				CreatedAt:      timestamppb.Now(),
				UpdatedAt:      timestamppb.Now(),
				EncryptionKeys: []*v1.DiaryEncryptionKey{{Id: "key-1", Value: envelope}},
			}}
			if putReceived {
				response = &v1.GetDiaryResponse{Diary: &v1.EncryptedDiary{
					Id:             "diary-1",
					CreatedAt:      timestamppb.Now(),
					UpdatedAt:      timestamppb.Now(),
					EncryptionKeys: []*v1.DiaryEncryptionKey{{Id: "key-1", Value: envelope}},
					Change:         req.GetEncryptedChange(),
				}}
			}
		case "io.thingsdiary.v1.diary.put":
			require.NoError(t, proto.Unmarshal(rpcPayload(t, method, data), &req))
			putReceived = true
			response = &v1.PutDiaryResponse{}
		default:
			t.Fatalf("unexpected RPC method: %s", method)
		}
		encoded, err := proto.Marshal(response)
		require.NoError(t, err)
		return centrifuge.RPCResult{Data: encoded}, nil
	}}

	// act
	diary, err := c.PutDiary(context.Background(), "diary-1", PutDiaryParams{
		Title:           "Updated diary",
		Description:     "Updated description",
		ParentChangeIDs: parents,
	})

	// assert
	require.NoError(t, err)
	require.Equal(t, []string{"io.thingsdiary.v1.diary.get", "io.thingsdiary.v1.diary.put", "io.thingsdiary.v1.diary.get"}, methods)
	require.Equal(t, "diary-1", req.GetDiaryId())
	require.Equal(t, "key-1", req.GetEncryptedChange().GetEncryption().GetDiaryKeyId())
	require.NotEmpty(t, req.GetEncryptedChange().GetEncryption().GetEncryptedKeyNonce())
	require.NotEmpty(t, req.GetEncryptedChange().GetEncryption().GetEncryptedKeyData())
	require.NotNil(t, req.GetEncryptedChange().GetChange())
	_, err = uuid.Parse(req.GetEncryptedChange().GetChange().GetId())
	require.NoError(t, err)
	require.Equal(t, "diary-1", req.GetEncryptedChange().GetChange().GetObjectId())
	require.Equal(t, parents, req.GetEncryptedChange().GetChange().GetParentChangeIds())
	require.Equal(t, req.GetEncryptedChange().GetChange().GetId(), diary.ChangeID)
	require.Equal(t, "Updated diary", diary.Title)
	require.Equal(t, "Updated description", diary.Description)

	nonce, encryptedData, err := UnmarshalChangePayload(req.GetEncryptedChange().GetChange().GetPayload())
	require.NoError(t, err)
	require.NotEmpty(t, nonce)
	require.NotEmpty(t, encryptedData)
	entityKey, err := decryptWithSymmetricKey(
		req.GetEncryptedChange().GetEncryption().GetEncryptedKeyNonce(),
		req.GetEncryptedChange().GetEncryption().GetEncryptedKeyData(),
		diaryKey,
	)
	require.NoError(t, err)
	plaintext, err := decryptWithSymmetricKey(nonce, encryptedData, entityKey)
	require.NoError(t, err)
	var details v1.DiaryDetails
	require.NoError(t, proto.Unmarshal(plaintext, &details))
	require.True(t, proto.Equal(&v1.DiaryDetails{Title: "Updated diary", Description: "Updated description"}, &details))
}

func TestUpdateDiary_RequiresParentsBeforeRPC(t *testing.T) {
	// arrange
	c := NewClient()
	c.credentials = &Credentials{}
	calls := 0
	c.conn = &mockRealtimeConn{state: centrifuge.StateConnected, rpc: func(context.Context, string, []byte) (centrifuge.RPCResult, error) {
		calls++
		return centrifuge.RPCResult{}, nil
	}}

	// act
	diary, err := c.PutDiary(context.Background(), "diary-1", PutDiaryParams{Title: "Updated diary"})

	// assert
	require.ErrorIs(t, err, ErrBadRequest)
	require.Nil(t, diary)
	require.Zero(t, calls)
}
