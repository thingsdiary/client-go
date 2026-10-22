package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/centrifugal/centrifuge-go"
	"github.com/samber/mo"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	v1 "github.com/thingsdiary/client-go/proto"
)

func TestGetRealtimeTokenUsesCurrentAuthToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/v1/realtime/token", r.URL.Path)
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "Bearer session-token", r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"token":"connection-token"}`))
	}))
	defer server.Close()

	c := NewClient(WithBaseURL(server.URL + "/api"))
	c.authToken = "session-token"
	token, err := c.GetRealtimeToken(context.Background())
	require.NoError(t, err)
	require.Equal(t, "connection-token", token)
}

func TestGetRealtimeTokenRequiresAuthentication(t *testing.T) {
	c := NewClient()
	_, err := c.GetRealtimeToken(context.Background())
	require.ErrorIs(t, err, ErrUnauthorized)
}

func TestEntryFromProtoDecryptsTransportToDomain(t *testing.T) {
	created := time.Date(2026, 10, 17, 12, 30, 0, 0, time.UTC)
	topicID := "topic-1"
	diaryKey := make([]byte, 32)
	entityKey := make([]byte, 32)
	detailsJSON := []byte(`{"content":"hello","archived":true}`)
	keyNonce, encryptedKey, err := encryptWithSymmetricKey(entityKey, diaryKey)
	require.NoError(t, err)
	detailsNonce, encryptedDetails, err := encryptWithSymmetricKey(detailsJSON, entityKey)
	require.NoError(t, err)
	entry := &v1.Entry{
		Id: "entry-1", DiaryId: "diary-1", Version: 42,
		CreatedAt: timestamppb.New(created), UpdatedAt: timestamppb.New(created.Add(time.Minute)),
		TopicId:    &topicID,
		Encryption: &v1.Encryption{DiaryKeyId: "key-1", EncryptedKeyNonce: keyNonce, EncryptedKeyData: encryptedKey},
		Details:    &v1.EncryptedData{Nonce: detailsNonce, Data: encryptedDetails},
	}
	got, err := entryFromProto(entry, diaryKey)
	require.NoError(t, err)
	require.Equal(t, "entry-1", got.ID)
	require.Equal(t, uint64(42), got.Version)
	require.Equal(t, created, got.CreatedAt)
	require.Equal(t, "topic-1", got.TopicID.MustGet())
	require.Equal(t, "hello", got.Content)
	require.True(t, got.Archived)
}

func TestRealtimeCodecPreservesOptionalString(t *testing.T) {
	require.Nil(t, protoString(mo.None[string]()))
	got := protoString(mo.Some("template-1"))
	require.NotNil(t, got)
	require.Equal(t, "template-1", *got)
}

func TestRealtimeEventDecodesResourceOneof(t *testing.T) {
	want := &v1.Event{Payload: &v1.Event_DiaryChanged{DiaryChanged: &v1.DiaryChanged{EventId: "event-1", Diary: &v1.Diary{Id: "diary-1"}}}}
	encoded, err := proto.Marshal(want)
	require.NoError(t, err)
	got := &v1.Event{}
	require.NoError(t, proto.Unmarshal(encoded, got))
	require.Equal(t, "diary-1", got.GetDiaryChanged().GetDiary().GetId())
	require.Equal(t, "event-1", got.GetDiaryChanged().GetEventId())
}

func TestRealtimeRPCDoesNotFallbackToHTTP(t *testing.T) {
	calledHTTP := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calledHTTP = true }))
	defer server.Close()
	c := NewClient(WithBaseURL(server.URL + "/api"))
	c.authToken = "session-token"
	_, err := c.getDiary(context.Background(), "diary-1")
	require.ErrorContains(t, err, "realtime client is not connected")
	require.False(t, calledHTTP)
}

func TestMapRealtimeRPCErrorPreservesPublicDomainErrors(t *testing.T) {
	tests := []struct {
		name   string
		method string
		code   uint32
		want   error
	}{
		{name: "unauthenticated", method: "io.thingsdiary.v1.diary.list", code: realtimeCodeUnauthenticated, want: ErrUnauthorized},
		{name: "permission denied", method: "io.thingsdiary.v1.entry.get", code: realtimeCodePermissionDenied, want: ErrForbidden},
		{name: "diary not found", method: "io.thingsdiary.v1.diary.get", code: realtimeCodeNotFound, want: ErrDiaryNotFound},
		{name: "entry not found", method: "io.thingsdiary.v1.entry.get", code: realtimeCodeNotFound, want: ErrEntryNotFound},
		{name: "topic not found", method: "io.thingsdiary.v1.topic.get", code: realtimeCodeNotFound, want: ErrTopicNotFound},
		{name: "template not found", method: "io.thingsdiary.v1.template.get", code: realtimeCodeNotFound, want: ErrTemplateNotFound},
		{name: "diary limit", method: "io.thingsdiary.v1.diary.create", code: realtimeCodeResourceExhausted, want: ErrDiaryLimitExceeded},
		{name: "invalid argument", method: "io.thingsdiary.v1.entry.put", code: realtimeCodeInvalidArgument, want: ErrBadRequest},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := mapRealtimeRPCError(test.method, &centrifuge.Error{Code: test.code, Message: "server detail"})
			require.ErrorIs(t, got, test.want)
		})
	}
}

func TestMapRealtimeRPCErrorLeavesTransportErrorsAlone(t *testing.T) {
	want := context.DeadlineExceeded
	require.ErrorIs(t, mapRealtimeRPCError("io.thingsdiary.v1.diary.get", want), want)
}

func TestMapRealtimeRPCFailedPreconditionToRemoved(t *testing.T) {
	got := mapRealtimeRPCError("io.thingsdiary.v1.entry.put", &centrifuge.Error{Code: realtimeCodeFailedPrecondition, Message: "failed precondition"})
	require.ErrorIs(t, got, ErrEntityRemoved)
}

func TestDeleteSoftValueDefaultsToSoftAndHonorsHardDelete(t *testing.T) {
	softTrue := true
	softFalse := false
	require.True(t, deleteSoftValue(nil))
	require.True(t, deleteSoftValue(&softTrue))
	require.False(t, deleteSoftValue(&softFalse))
}

func TestDeleteMethodsSendSoftDeleteMode(t *testing.T) {
	hardDelete := false

	tests := []struct {
		name       string
		method     string
		call       func(*Client) error
		assertData func(*testing.T, []byte)
	}{
		{
			name:   "entry defaults to soft delete",
			method: "io.thingsdiary.v1.entry.delete",
			call: func(c *Client) error {
				return c.DeleteEntry(context.Background(), "diary-1", "entry-1")
			},
			assertData: func(t *testing.T, data []byte) {
				var req v1.DeleteEntryRequest
				require.NoError(t, proto.Unmarshal(data, &req))
				require.Equal(t, "diary-1", req.GetDiaryId())
				require.Equal(t, "entry-1", req.GetEntryId())
				require.NotNil(t, req.Soft)
				require.True(t, req.GetSoft())
			},
		},
		{
			name:   "entry supports hard delete",
			method: "io.thingsdiary.v1.entry.delete",
			call: func(c *Client) error {
				return c.DeleteEntry(context.Background(), "diary-1", "entry-1", DeleteEntryParams{Soft: &hardDelete})
			},
			assertData: func(t *testing.T, data []byte) {
				var req v1.DeleteEntryRequest
				require.NoError(t, proto.Unmarshal(data, &req))
				require.NotNil(t, req.Soft)
				require.False(t, req.GetSoft())
			},
		},
		{
			name:   "topic defaults to soft delete",
			method: "io.thingsdiary.v1.topic.delete",
			call: func(c *Client) error {
				return c.DeleteTopic(context.Background(), "diary-1", "topic-1")
			},
			assertData: func(t *testing.T, data []byte) {
				var req v1.DeleteTopicRequest
				require.NoError(t, proto.Unmarshal(data, &req))
				require.Equal(t, "diary-1", req.GetDiaryId())
				require.Equal(t, "topic-1", req.GetTopicId())
				require.NotNil(t, req.Soft)
				require.True(t, req.GetSoft())
				require.False(t, req.GetDeleteEntries())
			},
		},
		{
			name:   "topic supports hard delete and entry deletion",
			method: "io.thingsdiary.v1.topic.delete",
			call: func(c *Client) error {
				return c.DeleteTopic(context.Background(), "diary-1", "topic-1", DeleteTopicParams{
					DeleteEntries: true,
					Soft:          &hardDelete,
				})
			},
			assertData: func(t *testing.T, data []byte) {
				var req v1.DeleteTopicRequest
				require.NoError(t, proto.Unmarshal(data, &req))
				require.NotNil(t, req.Soft)
				require.False(t, req.GetSoft())
				require.True(t, req.GetDeleteEntries())
			},
		},
		{
			name:   "template defaults to soft delete",
			method: "io.thingsdiary.v1.template.delete",
			call: func(c *Client) error {
				return c.DeleteTemplate(context.Background(), "diary-1", "template-1")
			},
			assertData: func(t *testing.T, data []byte) {
				var req v1.DeleteTemplateRequest
				require.NoError(t, proto.Unmarshal(data, &req))
				require.Equal(t, "diary-1", req.GetDiaryId())
				require.Equal(t, "template-1", req.GetTemplateId())
				require.NotNil(t, req.Soft)
				require.True(t, req.GetSoft())
			},
		},
		{
			name:   "template supports hard delete",
			method: "io.thingsdiary.v1.template.delete",
			call: func(c *Client) error {
				return c.DeleteTemplate(context.Background(), "diary-1", "template-1", DeleteTemplateParams{Soft: &hardDelete})
			},
			assertData: func(t *testing.T, data []byte) {
				var req v1.DeleteTemplateRequest
				require.NoError(t, proto.Unmarshal(data, &req))
				require.NotNil(t, req.Soft)
				require.False(t, req.GetSoft())
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := NewClient()
			c.credentials = &Credentials{SigningPrivateKey: testSigningPrivateKey()}
			c.conn = &mockRealtimeConn{
				state: centrifuge.StateConnected,
				rpc: func(_ context.Context, method string, data []byte) (centrifuge.RPCResult, error) {
					require.Equal(t, test.method, method)
					test.assertData(t, rpcPayload(t, method, data))
					return centrifuge.RPCResult{}, nil
				},
			}

			require.NoError(t, test.call(c))
		})
	}
}

func TestCloseFinishesEvents(t *testing.T) {
	conn := centrifuge.NewProtobufClient("ws://localhost:8000/connection/websocket", centrifuge.Config{})
	c := &Client{
		conn:   conn,
		events: make(chan *v1.Event),
	}
	events := c.Events()
	require.NotNil(t, events)
	c.Close()
	require.Nil(t, c.conn)
	require.Nil(t, c.Events())
	_, ok := <-events
	require.False(t, ok)
}
