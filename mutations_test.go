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

var (
	_ func(*Client, context.Context, CreateDiaryParams) (*Diary, error)                    = (*Client).CreateDiary
	_ func(*Client, context.Context, string, PutDiaryParams) (*Diary, error)               = (*Client).PutDiary
	_ func(*Client, context.Context, string, CreateEntryParams) (*Entry, error)            = (*Client).CreateEntry
	_ func(*Client, context.Context, string, string, PutEntryParams) (*Entry, error)       = (*Client).PutEntry
	_ func(*Client, context.Context, string, string) (*Entry, error)                       = (*Client).RestoreEntry
	_ func(*Client, context.Context, string, CreateTopicParams) (*Topic, error)            = (*Client).CreateTopic
	_ func(*Client, context.Context, string, string, PutTopicParams) (*Topic, error)       = (*Client).PutTopic
	_ func(*Client, context.Context, string, string) (*Topic, error)                       = (*Client).RestoreTopic
	_ func(*Client, context.Context, string, CreateTemplateParams) (*Template, error)      = (*Client).CreateTemplate
	_ func(*Client, context.Context, string, string, PutTemplateParams) (*Template, error) = (*Client).PutTemplate
	_ func(*Client, context.Context, string, string) (*Template, error)                    = (*Client).RestoreTemplate
)

func TestMutationsRequireCredentials(t *testing.T) {
	c := NewClient()
	ctx := context.Background()

	tests := []struct {
		name string
		call func() error
	}{
		{
			name: "create diary",
			call: func() error {
				_, err := c.CreateDiary(ctx, CreateDiaryParams{})
				return err
			},
		},
		{
			name: "put diary",
			call: func() error {
				_, err := c.PutDiary(ctx, "diary-1", PutDiaryParams{})
				return err
			},
		},
		{
			name: "create entry",
			call: func() error {
				_, err := c.CreateEntry(ctx, "diary-1", CreateEntryParams{})
				return err
			},
		},
		{
			name: "put entry",
			call: func() error {
				_, err := c.PutEntry(ctx, "diary-1", "entry-1", PutEntryParams{})
				return err
			},
		},
		{
			name: "restore entry",
			call: func() error {
				_, err := c.RestoreEntry(ctx, "diary-1", "entry-1")
				return err
			},
		},
		{
			name: "create topic",
			call: func() error {
				_, err := c.CreateTopic(ctx, "diary-1", CreateTopicParams{})
				return err
			},
		},
		{
			name: "put topic",
			call: func() error {
				_, err := c.PutTopic(ctx, "diary-1", "topic-1", PutTopicParams{})
				return err
			},
		},
		{
			name: "restore topic",
			call: func() error {
				_, err := c.RestoreTopic(ctx, "diary-1", "topic-1")
				return err
			},
		},
		{
			name: "create template",
			call: func() error {
				_, err := c.CreateTemplate(ctx, "diary-1", CreateTemplateParams{})
				return err
			},
		},
		{
			name: "put template",
			call: func() error {
				_, err := c.PutTemplate(ctx, "diary-1", "template-1", PutTemplateParams{})
				return err
			},
		},
		{
			name: "restore template",
			call: func() error {
				_, err := c.RestoreTemplate(ctx, "diary-1", "template-1")
				return err
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.ErrorIs(t, test.call(), ErrUnauthorized)
		})
	}
}

func TestCreateDiarySendsEncryptedRPCRequest(t *testing.T) {
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
		require.Equal(t, "io.thingsdiary.v1.diary.create", gotMethod)
		method = gotMethod
		require.NoError(t, proto.Unmarshal(rpcPayload(t, gotMethod, data), &req))
		now := timestamppb.Now()
		response := &v1.CreateDiaryResponse{
			Diary: &v1.EncryptedDiary{
				Id:        "diary-1",
				Version:   1,
				CreatedAt: now,
				UpdatedAt: now,
				EncryptionKeys: []*v1.DiaryEncryptionKey{{
					Id:    "key-1",
					Value: req.EncryptedDiaryKey,
				}},
				Encryption: &v1.Encryption{
					DiaryKeyId:        "key-1",
					EncryptedKeyNonce: req.EncryptedKeyNonce,
					EncryptedKeyData:  req.EncryptedKeyData,
				},
				Details: req.Details,
			},
		}
		encoded, err := proto.Marshal(response)
		require.NoError(t, err)
		return centrifuge.RPCResult{Data: encoded}, nil
	}}

	diary, err := c.CreateDiary(context.Background(), CreateDiaryParams{
		Title:       "Private diary",
		Description: "Encrypted description",
	})
	require.NoError(t, err)
	require.Equal(t, "io.thingsdiary.v1.diary.create", method)
	require.NotEmpty(t, req.GetEncryptedDiaryKey())
	require.NotEmpty(t, req.GetEncryptedKeyNonce())
	require.NotEmpty(t, req.GetEncryptedKeyData())
	require.NotEmpty(t, req.GetDetails().GetNonce())
	require.NotEmpty(t, req.GetDetails().GetData())
	require.NotNil(t, req.GetChange())
	_, err = uuid.Parse(req.GetChange().GetId())
	require.NoError(t, err)
	require.Equal(t, req.GetChange().GetId(), diary.ChangeID)
	require.Empty(t, req.GetChange().GetObjectId())
	require.Empty(t, req.GetChange().GetParentChangeIds())
	payloadNonce, payloadData, err := UnmarshalChangePayload(req.GetChange().GetPayload())
	require.NoError(t, err)
	require.Equal(t, req.GetDetails().GetNonce(), payloadNonce)
	require.Equal(t, req.GetDetails().GetData(), payloadData)

	diaryKey, ok := box.OpenAnonymous(nil, req.GetEncryptedDiaryKey(), publicKey, privateKey)
	require.True(t, ok)

	entityKey, err := decryptWithSymmetricKey(
		req.GetEncryptedKeyNonce(),
		req.GetEncryptedKeyData(),
		diaryKey,
	)
	require.NoError(t, err)

	detailsBytes, err := decryptWithSymmetricKey(
		req.GetDetails().GetNonce(),
		req.GetDetails().GetData(),
		entityKey,
	)
	require.NoError(t, err)

	var details v1.DiaryDetails
	require.NoError(t, proto.Unmarshal(detailsBytes, &details))
	require.True(t, proto.Equal(&v1.DiaryDetails{
		Title:       "Private diary",
		Description: "Encrypted description",
	}, &details))
}

func TestPutDiarySendsEncryptedChangeAndLegacyFields(t *testing.T) {
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
				EncryptionKeys: []*v1.DiaryEncryptionKey{{Id: "key-1", Value: envelope}},
			}}
		case "io.thingsdiary.v1.diary.put":
			require.NoError(t, proto.Unmarshal(rpcPayload(t, method, data), &req))
			response = &v1.PutDiaryResponse{Diary: &v1.EncryptedDiary{
				Id:             req.GetDiaryId(),
				Version:        req.GetVersion(),
				CreatedAt:      timestamppb.Now(),
				UpdatedAt:      timestamppb.Now(),
				EncryptionKeys: []*v1.DiaryEncryptionKey{{Id: "key-1", Value: envelope}},
				Encryption:     req.GetEncryption(),
				Details:        req.GetDetails(),
			}}
		default:
			t.Fatalf("unexpected RPC method: %s", method)
		}
		encoded, err := proto.Marshal(response)
		require.NoError(t, err)
		return centrifuge.RPCResult{Data: encoded}, nil
	}}

	diary, err := c.PutDiary(context.Background(), "diary-1", PutDiaryParams{
		Title:           "Updated diary",
		Description:     "Updated description",
		ParentChangeIDs: parents,
	})
	require.NoError(t, err)
	require.Equal(t, []string{"io.thingsdiary.v1.diary.get", "io.thingsdiary.v1.diary.put"}, methods)
	require.Equal(t, "diary-1", req.GetDiaryId())
	require.NotZero(t, req.GetVersion())
	require.Equal(t, "key-1", req.GetEncryption().GetDiaryKeyId())
	require.NotEmpty(t, req.GetEncryption().GetEncryptedKeyNonce())
	require.NotEmpty(t, req.GetEncryption().GetEncryptedKeyData())
	require.NotNil(t, req.GetChange())
	_, err = uuid.Parse(req.GetChange().GetId())
	require.NoError(t, err)
	require.Equal(t, "diary-1", req.GetChange().GetObjectId())
	require.Equal(t, parents, req.GetChange().GetParentChangeIds())
	require.Equal(t, []string{"head-a", "head-b"}, parents)
	require.Equal(t, req.GetChange().GetId(), diary.ChangeID)
	require.Equal(t, "Updated diary", diary.Title)
	require.Equal(t, "Updated description", diary.Description)

	nonce, encryptedData, err := UnmarshalChangePayload(req.GetChange().GetPayload())
	require.NoError(t, err)
	require.Equal(t, req.GetDetails().GetNonce(), nonce)
	require.Equal(t, req.GetDetails().GetData(), encryptedData)
	entityKey, err := decryptWithSymmetricKey(
		req.GetEncryption().GetEncryptedKeyNonce(),
		req.GetEncryption().GetEncryptedKeyData(),
		diaryKey,
	)
	require.NoError(t, err)
	plaintext, err := decryptWithSymmetricKey(nonce, encryptedData, entityKey)
	require.NoError(t, err)
	var details v1.DiaryDetails
	require.NoError(t, proto.Unmarshal(plaintext, &details))
	require.True(t, proto.Equal(&v1.DiaryDetails{Title: "Updated diary", Description: "Updated description"}, &details))
}

func TestPutDiaryRequiresParentsBeforeRPC(t *testing.T) {
	c := NewClient()
	c.credentials = &Credentials{}
	calls := 0
	c.conn = &mockRealtimeConn{state: centrifuge.StateConnected, rpc: func(context.Context, string, []byte) (centrifuge.RPCResult, error) {
		calls++
		return centrifuge.RPCResult{}, nil
	}}

	diary, err := c.PutDiary(context.Background(), "diary-1", PutDiaryParams{Title: "Updated diary"})
	require.ErrorIs(t, err, ErrBadRequest)
	require.Nil(t, diary)
	require.Zero(t, calls)
}
