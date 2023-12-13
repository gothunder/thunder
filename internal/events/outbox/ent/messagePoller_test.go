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

func TestNewEntMessagePoller(t *testing.T) {
	entClient := setupEnt(t)

	type args struct {
		outboxMessageClient interface{}
		pollInterval        time.Duration
		batchSize           int
	}
	tests := []struct {
		name    string
		args    args
		want    *entMessagePoller
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				outboxMessageClient: entClient.OutboxMessage,
				pollInterval:        1,
				batchSize:           1,
			},
			want: &entMessagePoller{
				client:       entClient.OutboxMessage,
				pollInterval: 1,
				batchSize:    1,
			},
			wantErr: false,
		},
		{
			name: "error",
			args: args{
				outboxMessageClient: nil,
				pollInterval:        1,
				batchSize:           1,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "error",
			args: args{
				outboxMessageClient: entClient.OutboxMessage,
				pollInterval:        0,
				batchSize:           1,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "error",
			args: args{
				outboxMessageClient: entClient.OutboxMessage,
				pollInterval:        1,
				batchSize:           0,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "error",
			args: args{
				outboxMessageClient: entClient,
				pollInterval:        1,
				batchSize:           1,
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewEntMessagePoller(tt.args.outboxMessageClient, tt.args.pollInterval, tt.args.batchSize)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewEntMessagePoller() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want == nil && got != nil {
				t.Errorf("NewEntMessagePoller() got = %v, want %v", got, tt.want)
			}
			if tt.want != nil && got == nil {
				t.Errorf("NewEntMessagePoller() got = %v, want %v", got, tt.want)
			}
			if tt.want != nil && got != nil {
				if got.(*entMessagePoller).client != tt.want.client {
					t.Errorf("NewEntMessagePoller() got client = %v, want %v", got, tt.want)
				}
				if got.(*entMessagePoller).pollInterval != tt.want.pollInterval {
					t.Errorf("NewEntMessagePoller() got pollInterval = %v, want %v", got, tt.want)
				}
				if got.(*entMessagePoller).batchSize != tt.want.batchSize {
					t.Errorf("NewEntMessagePoller() got batchSize = %v, want %v", got, tt.want)
				}
				close(got.(*entMessagePoller).closeChan)
				close(got.(*entMessagePoller).nextChan)
			}
		})
	}
}

func TestEntMessagePoller_Close(t *testing.T) {
	t.Run("polling already closed", func(t *testing.T) {
		// Arrange
		entClient := setupEnt(t)
		e, err := NewEntMessagePoller(entClient.OutboxMessage, 1, 1)
		if err != nil {
			t.Errorf("NewEntMessagePoller error = %v, wantErr %v", err, nil)
			return
		}
		e = outbox.WrapPollerWithTracing(e)

		// Act
		if err := e.Close(); err != nil {
			t.Errorf("entMessagePoller.Close() error = %v, wantErr %v", err, nil)
		}
		messages, _, err := e.Poll(context.Background())

		if messages != nil {
			t.Errorf("entMessagePoller.Close() messages = %v, want %v", messages, nil)
		}
		if err == nil {
			t.Errorf("entMessagePoller.Close() error = %v, wantErr %v", err, ErrMessagePollerClosed)
		}
	})

	t.Run("close a polling", func(t *testing.T) {
		// Arrange
		entClient := setupEnt(t)
		e, err := NewEntMessagePoller(entClient.OutboxMessage, 1, 1)
		if err != nil {
			t.Errorf("NewEntMessagePoller error = %v, wantErr %v", err, nil)
			return
		}
		e = outbox.WrapPollerWithTracing(e)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		messagesChan, _, err := e.Poll(ctx)
		if err != nil {
			t.Errorf("entMessagePoller.Close() error = %v, wantErr %v", err, nil)
			return
		}

		// Act
		if err := e.Close(); err != nil {
			t.Errorf("entMessagePoller.Close() error = %v, wantErr %v", err, nil)
		}
		messages := <-messagesChan

		if messages != nil {
			t.Errorf("entMessagePoller.Close() messages = %v, want %v", messages, nil)
		}
	})
}

func TestEntMessagePoller_Poll(t *testing.T) {
	// Arrange
	population := []entInit.OutboxMessage{
		{
			ID:          uuid.MustParse("6d1559ea-4a68-4c10-9646-b1a42cb9c6cd"),
			Payload:     []byte("payload"),
			Topic:       "delivered",
			CreatedAt:   time.Now(),
			DeliveredAt: time.Now(),
		},
		{
			ID:        uuid.MustParse("592ea815-34a9-4924-9097-82e59476f14a"),
			Payload:   []byte("payload"),
			Topic:     "notDelivered",
			CreatedAt: time.Now(),
		},
	}
	insertion := []entInit.OutboxMessage{
		{
			ID:        uuid.MustParse("6d5b845b-dd40-41f8-a6d6-b43639697abc"),
			Payload:   []byte("payload"),
			Topic:     "notDeliveredToo",
			CreatedAt: time.Now(),
		},
	}

	entClient := setupEnt(t)
	err := populateOutboxMessages(context.Background(), entClient, population)
	if err != nil {
		t.Errorf("populateOutboxMessages error = %v, wantErr %v", err, nil)
		return
	}

	wantMessages := []*outbox.Message{
		{
			ID:        population[1].ID,
			Payload:   population[1].Payload,
			Topic:     population[1].Topic,
			CreatedAt: population[1].CreatedAt,
		},
		{
			ID:        insertion[0].ID,
			Payload:   insertion[0].Payload,
			Topic:     insertion[0].Topic,
			CreatedAt: insertion[0].CreatedAt,
		},
	}

	e, err := NewEntMessagePoller(entClient.OutboxMessage, 1, 1)
	if err != nil {
		t.Errorf("NewEntMessagePoller error = %v, wantErr %v", err, nil)
		return
	}
	e = outbox.WrapPollerWithTracing(e)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Act
	messagesChan, next, err := e.Poll(ctx)

	messages := make([]*outbox.Message, 0)
	go func() {
		for msgPack := range messagesChan {
			messages = append(messages, msgPack...)
			ids := make([]uuid.UUID, len(msgPack))
			for i, msg := range msgPack {
				ids[i] = msg.ID
			}
			entClient.OutboxMessage.Update().SetDeliveredAt(time.Now()).Where(outboxmessage.IDIn(ids...)).ExecX(ctx)
			next()
		}
	}()

	time.Sleep(10 * time.Millisecond)
	err = populateOutboxMessages(context.Background(), entClient, insertion)
	if err != nil {
		t.Errorf("populateOutboxMessages error = %v, wantErr %v", err, nil)
		return
	}
	time.Sleep(10 * time.Millisecond)
	cancel()

	// Assert
	if messages == nil {
		t.Errorf("entMessagePoller.Poll() messages = %v, want %v", messages, wantMessages)
	}
	if len(messages) != len(wantMessages) {
		t.Errorf("entMessagePoller.Poll() len messages = %v, want %v", len(messages), len(wantMessages))
		return
	}
	for i := range messages {
		if messages[i].ID != wantMessages[i].ID {
			t.Errorf("entMessagePoller.Poll() message.ID = %v, want %v", messages[i].ID, wantMessages[i].ID)
		}
		if string(messages[i].Payload) != string(wantMessages[i].Payload) {
			t.Errorf("entMessagePoller.Poll() message.Payload = %v, want %v", string(messages[i].Payload), string(wantMessages[i].Payload))
		}
		if messages[i].Topic != wantMessages[i].Topic {
			t.Errorf("entMessagePoller.Poll() message.Topic = %v, want %v", messages[i].Topic, wantMessages[i].Topic)
		}
		if !messages[i].CreatedAt.Equal(wantMessages[i].CreatedAt) {
			t.Errorf("entMessagePoller.Poll() message.CreatedAt = %v, want %v", messages[i].CreatedAt, wantMessages[i].CreatedAt)
		}
	}
}
