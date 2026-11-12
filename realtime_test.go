package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/centrifugal/centrifuge-go"
	"github.com/samber/mo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "github.com/thingsdiary/client-go/proto"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
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
	// arrange
	created := time.Date(2026, 10, 17, 12, 30, 0, 0, time.UTC)
	removed := created.Add(2 * time.Minute)
	topicID := "topic-1"
	diaryKey := make([]byte, 32)
	entityKey := make([]byte, 32)
	detailsBytes, err := proto.Marshal(&v1.EntryDetails{Content: "hello", Archived: true})
	require.NoError(t, err)
	keyNonce, encryptedKey, err := encryptWithSymmetricKey(entityKey, diaryKey)
	require.NoError(t, err)
	detailsNonce, encryptedDetails, err := encryptWithSymmetricKey(detailsBytes, entityKey)
	require.NoError(t, err)
	payload, err := MarshalChangePayload(detailsNonce, encryptedDetails)
	require.NoError(t, err)
	entry := &v1.EncryptedEntry{
		Id: "entry-1", DiaryId: "diary-1",
		CreatedAt: timestamppb.New(created), UpdatedAt: timestamppb.New(created.Add(time.Minute)),
		RemovedAt: timestamppb.New(removed),
		TopicId:   &topicID,
		Change: &v1.EncryptedChange{
			Change:     &v1.Change{Id: "change-42", ObjectId: "entry-1", Payload: payload},
			Encryption: &v1.Encryption{DiaryKeyId: "key-1", EncryptedKeyNonce: keyNonce, EncryptedKeyData: encryptedKey},
		},
	}

	// act
	got, err := entryFromProto(entry, diaryKey)

	// assert
	require.NoError(t, err)
	require.Equal(t, "entry-1", got.ID)
	require.Equal(t, "change-42", got.ChangeID)
	require.Equal(t, created, got.CreatedAt)
	require.Equal(t, "topic-1", got.TopicID.MustGet())
	require.Equal(t, "hello", got.Content)
	require.True(t, got.Archived)
	assert.False(t, got.DeletedAt.IsPresent())
	assert.Equal(t, mo.Some(removed), got.RemovedAt)
}

func TestTopicFromProtoDecryptsChangeToDomain(t *testing.T) {
	// arrange
	created := time.Date(2026, 10, 17, 12, 30, 0, 0, time.UTC)
	diaryKey := make([]byte, 32)
	entityKey := make([]byte, 32)
	detailsBytes, err := proto.Marshal(&v1.TopicDetails{Title: "work", Color: "#123456"})
	require.NoError(t, err)
	keyNonce, encryptedKey, err := encryptWithSymmetricKey(entityKey, diaryKey)
	require.NoError(t, err)
	detailsNonce, encryptedDetails, err := encryptWithSymmetricKey(detailsBytes, entityKey)
	require.NoError(t, err)
	payload, err := MarshalChangePayload(detailsNonce, encryptedDetails)
	require.NoError(t, err)
	topic := &v1.EncryptedTopic{
		Id: "topic-1", DiaryId: "diary-1",
		CreatedAt: timestamppb.New(created), UpdatedAt: timestamppb.New(created.Add(time.Minute)),
		Change: &v1.EncryptedChange{
			Change:     &v1.Change{Id: "change-7", ObjectId: "topic-1", Payload: payload},
			Encryption: &v1.Encryption{DiaryKeyId: "key-1", EncryptedKeyNonce: keyNonce, EncryptedKeyData: encryptedKey},
		},
	}

	// act
	got, err := topicFromProto(topic, diaryKey)

	// assert
	require.NoError(t, err)
	require.Equal(t, "topic-1", got.ID)
	require.Equal(t, "change-7", got.ChangeID)
	require.Equal(t, "work", got.Title)
	require.Equal(t, "#123456", got.Color)
}

func TestRealtimeCodecPreservesOptionalString(t *testing.T) {
	require.Nil(t, protoString(mo.None[string]()))
	got := protoString(mo.Some("template-1"))
	require.NotNil(t, got)
	require.Equal(t, "template-1", *got)
}

func TestRealtimeEventDecodesResourceOneof(t *testing.T) {
	want := &v1.Event{Payload: &v1.Event_DiaryChanged{DiaryChanged: &v1.DiaryChanged{EventId: "event-1", DiaryId: "diary-1"}}}
	encoded, err := proto.Marshal(want)
	require.NoError(t, err)
	got := &v1.Event{}
	require.NoError(t, proto.Unmarshal(encoded, got))
	require.Equal(t, "diary-1", got.GetDiaryChanged().GetDiaryId())
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

func TestMapDiaryErrorCode(t *testing.T) {
	t.Parallel()
	// arrange
	tests := []struct {
		name string
		code v1.ErrorCode
		want error
	}{
		{name: "success", code: v1.ErrorCode_ERROR_CODE_UNSPECIFIED},
		{name: "bad data", code: v1.ErrorCode_ERROR_CODE_BAD_DATA, want: ErrBadRequest},
		{name: "diary limit", code: v1.ErrorCode_ERROR_CODE_DIARIES_LIMIT_REACHED, want: ErrDiaryLimitExceeded},
		{name: "diary not found", code: v1.ErrorCode_ERROR_CODE_NOT_FOUND, want: ErrDiaryNotFound},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			// act
			err := mapDiaryErrorCode(test.code)

			// assert
			if test.want == nil {
				require.NoError(t, err)
				return
			}
			require.ErrorIs(t, err, test.want)
		})
	}
}

func TestMapEntryErrorCode(t *testing.T) {
	t.Parallel()

	// arrange
	tests := []struct {
		name string
		code v1.ErrorCode
		want error
	}{
		{name: "success", code: v1.ErrorCode_ERROR_CODE_UNSPECIFIED},
		{name: "bad data", code: v1.ErrorCode_ERROR_CODE_BAD_DATA, want: ErrBadRequest},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// act
			err := mapEntryErrorCode(test.code)

			// assert
			if test.want == nil {
				require.NoError(t, err)

				return
			}
			require.ErrorIs(t, err, test.want)
		})
	}
}

func TestMapTopicErrorCode(t *testing.T) {
	t.Parallel()

	// arrange
	tests := []struct {
		name string
		code v1.ErrorCode
		want error
	}{
		{name: "success", code: v1.ErrorCode_ERROR_CODE_UNSPECIFIED},
		{name: "bad data", code: v1.ErrorCode_ERROR_CODE_BAD_DATA, want: ErrBadRequest},
		{name: "entry changes required", code: v1.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, want: ErrTopicEntriesChangeRequired},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// act
			err := mapTopicErrorCode(test.code)

			// assert
			if test.want == nil {
				require.NoError(t, err)

				return
			}
			require.ErrorIs(t, err, test.want)
		})
	}
}

func TestMapRealtimeRPCFailedPreconditionToRemoved(t *testing.T) {
	got := mapRealtimeRPCError("io.thingsdiary.v1.entry.put", &centrifuge.Error{Code: realtimeCodeFailedPrecondition, Message: "failed precondition"})
	require.ErrorIs(t, got, ErrEntityRemoved)
}

func TestDeleteMethodsSendSoftDeleteMode(t *testing.T) {
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
				return c.DeleteEntry(context.Background(), "diary-1", "entry-1", DeleteEntryParams{Soft: mo.Some(false)})
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
					Soft:          mo.Some(false),
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
				return c.DeleteTemplate(context.Background(), "diary-1", "template-1", DeleteTemplateParams{Soft: mo.Some(false)})
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
					if test.method == "io.thingsdiary.v1.entry.delete" && method == "io.thingsdiary.v1.entry.get" {
						response, err := proto.Marshal(&v1.GetEntryResponse{Entry: testEncryptedEntry()})
						require.NoError(t, err)

						return centrifuge.RPCResult{Data: response}, nil
					}
					if test.method == "io.thingsdiary.v1.topic.delete" && method == "io.thingsdiary.v1.topic.get" {
						response, err := proto.Marshal(&v1.GetTopicResponse{Topic: testEncryptedTopic()})
						require.NoError(t, err)

						return centrifuge.RPCResult{Data: response}, nil
					}

					require.Equal(t, test.method, method)
					test.assertData(t, rpcPayload(t, method, data))
					return centrifuge.RPCResult{}, nil
				},
			}

			require.NoError(t, test.call(c))
		})
	}
}

func testEncryptedEntry() *v1.EncryptedEntry {
	return &v1.EncryptedEntry{
		Id:      "entry-1",
		DiaryId: "diary-1",
		Change: &v1.EncryptedChange{
			Change: &v1.Change{
				Id:       "change-1",
				ObjectId: "entry-1",
				Payload:  []byte("payload"),
			},
			Encryption: &v1.Encryption{DiaryKeyId: "key-1"},
		},
	}
}

func testEncryptedTopic() *v1.EncryptedTopic {
	return &v1.EncryptedTopic{
		Id:      "topic-1",
		DiaryId: "diary-1",
		Change: &v1.EncryptedChange{
			Change: &v1.Change{
				Id:       "change-1",
				ObjectId: "topic-1",
				Payload:  []byte("payload"),
			},
			Encryption: &v1.Encryption{DiaryKeyId: "key-1"},
		},
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
