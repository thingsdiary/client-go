package client

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"testing"

	"github.com/centrifugal/centrifuge-go"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/nacl/box"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	v1 "github.com/thingsdiary/client-go/proto"
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
		response := &v1.Diary{
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
		}
		encoded, err := proto.Marshal(response)
		require.NoError(t, err)
		return centrifuge.RPCResult{Data: encoded}, nil
	}}

	_, err = c.CreateDiary(context.Background(), CreateDiaryParams{
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

	diaryKey, ok := box.OpenAnonymous(nil, req.GetEncryptedDiaryKey(), publicKey, privateKey)
	require.True(t, ok)

	entityKey, err := decryptWithSymmetricKey(
		req.GetEncryptedKeyNonce(),
		req.GetEncryptedKeyData(),
		diaryKey,
	)
	require.NoError(t, err)

	detailsJSON, err := decryptWithSymmetricKey(
		req.GetDetails().GetNonce(),
		req.GetDetails().GetData(),
		entityKey,
	)
	require.NoError(t, err)

	var details DiaryDetails
	require.NoError(t, json.Unmarshal(detailsJSON, &details))
	require.Equal(t, DiaryDetails{
		Title:       "Private diary",
		Description: "Encrypted description",
	}, details)
}
