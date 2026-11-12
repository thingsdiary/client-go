package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
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

func TestMutationAuthorization_RequiresCredentials(t *testing.T) {
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
			// arrange

			// act
			err := test.call()

			// assert
			require.ErrorIs(t, err, ErrUnauthorized)
		})
	}
}
