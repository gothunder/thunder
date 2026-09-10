package consumer

import (
	"context"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/gothunder/thunder/internal/events/rabbitmq"
	"github.com/gothunder/thunder/internal/events/rabbitmq/manager"
	"github.com/gothunder/thunder/pkg/events"
	"github.com/rotisserie/eris"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

type stubHandler struct{}

func (stubHandler) Topics() []string { return []string{"test.topic"} }

func (stubHandler) Handle(context.Context, string, events.EventDecoder) events.HandlerResponse {
	return events.Success
}

func testConsumer(t *testing.T) *rabbitmqConsumer {
	t.Helper()

	logger := zerolog.New(io.Discard)

	return &rabbitmqConsumer{
		config: rabbitmq.Config{ConsumerConcurrency: 1},
		logger: &logger,

		chManager: &manager.ChannelManager{
			ChannelMux:         &sync.RWMutex{},
			NotifyReconnection: make(chan error),
		},

		wg:        &sync.WaitGroup{},
		backoffWg: &sync.WaitGroup{},
	}
}

// The broker can refuse the registration right after a reconnection it just
// accepted. A pod that gave up there stayed alive and ready while consuming
// nothing, which is worse than dying.
func TestSubscribeRetriesTheRegistration(t *testing.T) {
	consumer := testConsumer(t)

	var attempts int
	registered := make(chan struct{})
	consumer.startGoRoutinesFunc = func(events.Handler) error {
		attempts++
		if attempts == 1 {
			return eris.New(`Exception (503) Reason: "unexpected command received"`)
		}

		close(registered)
		return nil
	}

	subscribed := make(chan error, 1)
	go func() {
		subscribed <- consumer.Subscribe(context.Background(), stubHandler{})
	}()

	select {
	case <-registered:
	case err := <-subscribed:
		t.Fatalf("the subscription died on a retryable failure: %v", err)
	case <-time.After(5 * time.Second):
		t.Fatal("the registration was never retried")
	}

	// The subscription is now waiting for a reconnection, so a failed one is
	// what finally ends it.
	consumer.chManager.NotifyReconnection <- eris.New("failed to reconnect to amqp")

	select {
	case err := <-subscribed:
		require.ErrorContains(t, err, "failed to reconnect to the amqp channel")
		require.Equal(t, 2, attempts, "the registration must be retried exactly once")
	case <-time.After(5 * time.Second):
		t.Fatal("the subscription never returned")
	}
}

// The retry loop must not delay a graceful shutdown, and the error it returns
// has to be recognizable as a cancellation so the fx wrapper doesn't report it
// as a failure and tear the application down.
func TestSubscribeStopsRetryingWhenTheContextIsCancelled(t *testing.T) {
	consumer := testConsumer(t)
	consumer.startGoRoutinesFunc = func(events.Handler) error {
		return eris.New("broker is down")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	subscribed := make(chan error, 1)
	go func() {
		subscribed <- consumer.Subscribe(ctx, stubHandler{})
	}()

	select {
	case err := <-subscribed:
		require.ErrorIs(t, err, context.Canceled)
	case <-time.After(200 * time.Millisecond):
		t.Fatal("the retry loop ignored the cancelled context")
	}
}
