package outboxent

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gothunder/thunder/internal/events/outbox"
	"github.com/gothunder/thunder/internal/events/outbox/ent/entInit"
	"github.com/gothunder/thunder/internal/events/outbox/ent/entInit/outboxmessage"
)

func TestNewEntMessageMarker(t *testing.T) {
	entClient := setupEnt(t)

	type args struct {
		outboxMessageClient interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				outboxMessageClient: entClient.OutboxMessage,
			},
			wantErr: false,
		},
		{
			name: "error",
			args: args{
				outboxMessageClient: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewEntMessageMarker(tt.args.outboxMessageClient)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewEntMessageMarker() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && got != nil {
				t.Errorf("NewEntMessageMarker() got = %v, want nil", got)
			}
		})
	}
}

func TestEntMessageMarker_MarkAsPublished(t *testing.T) {
	type fields struct {
		client *entInit.Client
	}
	type args struct {
		ctx     context.Context
		msgPack []*outbox.Message
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		popuation []entInit.OutboxMessage
		wantErr   bool
	}{
		{
			name: "success",
			fields: fields{
				client: setupEnt(t),
			},
			args: args{
				ctx: context.Background(),
				msgPack: []*outbox.Message{
					{
						ID: uuid.MustParse("14d8e114-71c0-4309-81aa-351d64dd9d74"),
					},
					{
						ID: uuid.MustParse("934997bd-eee8-4e1f-810a-4fd601ad8b9c"),
					},
				},
			},
			popuation: []entInit.OutboxMessage{
				{
					ID:      uuid.MustParse("14d8e114-71c0-4309-81aa-351d64dd9d74"),
					Topic:   "topic",
					Payload: []byte("payload"),
					Headers: map[string]string{
						"key": "value",
					},
					CreatedAt:   time.Now(),
					DeliveredAt: time.Now(),
				},
				{
					ID:      uuid.MustParse("934997bd-eee8-4e1f-810a-4fd601ad8b9c"),
					Topic:   "topic",
					Payload: []byte("payload"),
					Headers: map[string]string{
						"key": "value",
					},
					CreatedAt: time.Now(),
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			populateOutboxMessages(tt.args.ctx, tt.fields.client, tt.popuation)
			marker, err := NewEntMessageMarker(tt.fields.client.OutboxMessage)
			if err != nil {
				t.Fatal(err)
			}

			if err := marker.MarkAsPublished(tt.args.ctx, tt.args.msgPack); (err != nil) != tt.wantErr {
				t.Errorf("EntMessageMarker.MarkAsPublished() error = %v, wantErr %v", err, tt.wantErr)
			}

			// check if the messages were marked as delivered
			for _, msg := range tt.args.msgPack {
				entMsg, err := tt.fields.client.OutboxMessage.
					Query().
					Where(outboxmessage.ID(msg.ID)).
					First(context.Background())
				if err != nil {
					t.Fatal(err)
				}

				if entMsg.DeliveredAt.IsZero() {
					t.Errorf("EntMessageMarker.MarkAsPublished() message %s was not marked as delivered", msg.ID)
				}
			}
		})
	}
}
