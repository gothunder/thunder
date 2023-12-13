package outbox

import (
	"context"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/rs/zerolog"
)

type withLoggingReayer struct {
	next Relayer
}

// Close implements Relayer.
func (w *withLoggingReayer) Close(ctx context.Context) error {
	logger := zerolog.Ctx(ctx)

	logger.Debug().
		Str("context", "thunder.outbox.relayer").
		Str("op", "Close").Msg("closing relayer")

	defer func() {
		logger.Info().
			Str("context", "thunder.outbox.relayer").
			Str("op", "Close").Msg("reayer closed")
	}()

	return w.next.Close(ctx)
}

// Start implements Relayer.
func (w *withLoggingReayer) Start(ctx context.Context) (err error) {
	logger := zerolog.Ctx(ctx)

	logger.Debug().
		Str("context", "thunder.outbox.relayer").
		Str("op", "Start").Msg("starting relaying messages")

	defer func() {
		if err != nil {
			logger.Error().
				Str("context", "thunder.outbox.relayer").
				Str("op", "Start").
				Err(err).
				Msg("relaying messages failed to start")
		} else {
			logger.Info().
				Str("context", "thunder.outbox.relayer").
				Str("op", "Start").Msg("relaying messages started")
		}
	}()

	err = w.next.Start(ctx)
	return
}

// prepareMessages implements Relayer.
func (w *withLoggingReayer) prepareMessages(msgPack []*Message) []*message.Message {
	return w.next.prepareMessages(msgPack)
}

// relay implements Relayer.
func (w *withLoggingReayer) relay(ctx context.Context, msgPack []*Message) (err error) {
	logger := zerolog.Ctx(ctx)
	start := time.Now()

	logger.Debug().
		Str("context", "thunder.outbox.relayer").
		Str("op", "relay").
		Msg("starting relaying messages")

	defer func() {
		if err != nil {
			logger.Info().
				Str("context", "thunder.outbox.relayer").
				Str("op", "relay").
				Int("messages_num", len(msgPack)).
				Dur("latency", time.Since(start)).
				Err(err).
				Msg("relaying messages failed... it will be retried")
		} else {
			logger.Info().
				Str("context", "thunder.outbox.relayer").
				Str("op", "relay").
				Int("messages_num", len(msgPack)).
				Dur("latency", time.Since(start)).
				Msg("messages relayed")
		}
	}()

	err = w.next.relay(ctx, msgPack)
	return
}

func WrapRelayerWithLogging(next Relayer) Relayer {
	return &withLoggingReayer{
		next: next,
	}
}
