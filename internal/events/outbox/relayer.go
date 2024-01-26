package outbox

import (
	"context"
	"errors"
	"time"

	"github.com/TheRafaBonin/roxy"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/cenkalti/backoff/v4"
)

var (
	ErrRelayerClosed = errors.New("relayer is closed")
	ErrNilBackoff    = errors.New("backoff is nil")
	ErrNilPublisher  = errors.New("publisher is nil")
	ErrNilPoller     = errors.New("poller is nil")
	ErrNilMarker     = errors.New("marker is nil")
)

type Relayer interface {
	Start(ctx context.Context) error
	Close(ctx context.Context) error
	relay(ctx context.Context, msgPack []*Message) error
	prepareMessages(msgPack []*Message) []*message.Message
}

type MessagePoller interface {
	Poll(ctx context.Context) (<-chan []*Message, func(), error)
	Close() error
}

type MessageMarker interface {
	MarkAsPublished(ctx context.Context, msgPack []*Message) error
}

type relayer struct {
	publisher message.Publisher
	poller    MessagePoller
	marker    MessageMarker
	backOff   backoff.BackOff

	closedChan chan struct{}
}

func NewRelayer(
	backOff backoff.BackOff,
	publisher message.Publisher,
	poller MessagePoller,
	marker MessageMarker,
) (Relayer, error) {
	if backOff == nil {
		return nil, roxy.Wrap(ErrNilBackoff, "creating relayer")
	}

	if publisher == nil {
		return nil, roxy.Wrap(ErrNilPublisher, "creating relayer")
	}

	if poller == nil {
		return nil, roxy.Wrap(ErrNilPoller, "creating relayer")
	}

	if marker == nil {
		return nil, roxy.Wrap(ErrNilMarker, "creating relayer")
	}

	return &relayer{
		backOff:    backOff,
		poller:     poller,
		marker:     marker,
		publisher:  publisher,
		closedChan: make(chan struct{}),
	}, nil
}

func (r *relayer) Start(ctx context.Context) error {
	msgPacks, next, _ := r.poller.Poll(ctx)

	for msgPack := range msgPacks {
		select {
		case <-r.closedChan:
			return ErrRelayerClosed
		case <-ctx.Done():
			return ctx.Err()
		default:
			err := backoff.Retry(func() error {
				return r.relay(ctx, msgPack)
			}, backoff.WithContext(r.backOff, ctx))

			if err != nil {
				return roxy.Wrap(err, "relaying messages")
			}

			next()
		}
	}

	return nil
}

func (r *relayer) relay(ctx context.Context, msgPack []*Message) error {
	select {
	case <-r.closedChan:
		return backoff.Permanent(roxy.Wrap(ErrRelayerClosed, "relaying messages"))
	default:
		msgs := r.prepareMessages(msgPack)

		for i, msg := range msgs {
			// skips messages that have already been delivered in failed attempts
			if !msgPack[i].DeliveredAt.IsZero() {
				continue
			}

			// publishes one at a time to prevent desordering
			if err := r.publisher.Publish(msgPack[i].Topic, msg); err != nil {
				// marks all previous messages as delivered in case of error to prevent
				// message duplication
				if markErr := r.marker.MarkAsPublished(ctx, msgPack[:i]); markErr != nil {
					return roxy.Wrap(markErr, "marking messages as published")
				}
				return roxy.Wrap(err, "publishing messages")
			}

			// marks message as delivered
			msgPack[i].DeliveredAt = time.Now()
		}

		// marks the whole pack as delivered
		if err := r.marker.MarkAsPublished(ctx, msgPack); err != nil {
			return roxy.Wrap(err, "marking messages as published")
		}

		return nil
	}
}

func (r *relayer) prepareMessages(msgPack []*Message) []*message.Message {
	msgs := make([]*message.Message, 0, len(msgPack))

	for _, msg := range msgPack {
		wMsg := message.NewMessage(msg.ID.String(), msg.Payload)
		if msg.Headers != nil {
			wMsg.Metadata = msg.Headers
		}

		msgs = append(msgs, wMsg)
	}

	return msgs
}

func (r *relayer) Close(ctx context.Context) error {
	close(r.closedChan)
	errs := make([]error, 2)
	errs[0] = r.poller.Close()
	errs[1] = r.publisher.Close()

	return errors.Join(errs...)
}
