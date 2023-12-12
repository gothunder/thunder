package outbox

import (
	"context"
	"errors"

	"github.com/TheRafaBonin/roxy"
)

var (
	ErrNoMessages = errors.New("no messages")
)

type Storer interface {
	Store(ctx context.Context, txOutboxMessageClient interface{}, messages []Message) error
	WithTxClient(txOutboxMessageClient interface{}) (TransactionalStorer, error)
}

type TransactionalStorer interface {
	Store(ctx context.Context, messages []Message) error
}

type storer struct{}

// Store implements Storer.
func (s storer) Store(ctx context.Context, txOutboxMessageClient interface{}, messages []Message) error {
	if err := validateMessages(messages); err != nil {
		return roxy.Wrap(err, "validating messages")
	}

	txClient, err := WrapOutboxMessageClient(txOutboxMessageClient)
	if err != nil {
		return roxy.Wrap(err, "wrapping tx client")
	}

	entMessages := make([]MessageCreator, len(messages))
	for i, msg := range messages {
		entMessages[i] = msg.BuildEntMessage(txClient.Create())
	}

	err = txClient.CreateBulk(entMessages...).Exec(ctx)
	if err != nil {
		return roxy.Wrap(err, "creating messages")
	}

	return nil
}

// WithTxClient implements Storer.
func (s *storer) WithTxClient(txOutboxMessageClient interface{}) (TransactionalStorer, error) {
	return newTransactionalStorer(s, txOutboxMessageClient)
}

func NewStorer() Storer {
	return &storer{}
}

type transactionalStorer struct {
	storer   Storer
	txClient interface{}
}

// Store implements TransactionalStorer.
func (t transactionalStorer) Store(ctx context.Context, messages []Message) error {
	return t.storer.Store(ctx, t.txClient, messages)
}

func newTransactionalStorer(storer Storer, txOutboxMessageClient interface{}) (TransactionalStorer, error) {
	if err := validateOutboxMessageClient(txOutboxMessageClient); err != nil {
		return nil, roxy.Wrap(err, "validating tx client")
	}

	return &transactionalStorer{
		storer:   storer,
		txClient: txOutboxMessageClient,
	}, nil
}

func validateMessages(messages []Message) error {
	if len(messages) == 0 {
		return ErrNoMessages
	}

	errs := make([]error, len(messages))
	for i, msg := range messages {
		errs[i] = msg.Validate()
	}
	return errors.Join(errs...)
}
