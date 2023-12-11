package outbox

import (
	"context"
	"time"

	"github.com/rs/zerolog"
)

const (
	opLabel          = "op"
	op               = "thunder.outbox.storer.Store"
	latencyLabel     = "latency"
	messagesNumLabel = "messages_num"
)

type withLoggingStorer struct {
	next Storer
}

// Store implements Storer.
func (wts withLoggingStorer) Store(ctx context.Context, tx interface{}, messages []Message) error {
	logger := zerolog.Ctx(ctx)
	logger.Debug().Str(opLabel, op).Msg("starting storing messages")

	start := time.Now()

	defer func() {
		logger.Debug().Str(opLabel, op).Msg("finished storing messages")
		logger.
			Info().
			Dur(latencyLabel, time.Since(start)).
			Str(opLabel, op).
			Int(messagesNumLabel, len(messages)).
			Msg("messages stored")
	}()

	return wts.next.Store(ctx, tx, messages)
}

// WithTxClient implements Storer.
func (wts *withLoggingStorer) WithTxClient(tx interface{}) (TransactionalStorer, error) {
	// needs to be reimplemented here or otherwise TransactionalStorer
	// will not have logging
	return newTransactionalStorer(wts, tx)
}

func WrapStorerWithLogging(next Storer) Storer {
	return &withLoggingStorer{
		next: next,
	}
}
