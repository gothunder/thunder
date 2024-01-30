package outboxpublisher

import (
	"context"
	"slices"
	"testing"

	"github.com/ThreeDotsLabs/watermill/components/forwarder"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/gothunder/thunder/internal/events/rabbitmq/tracing"
	"github.com/rotisserie/eris"
)

type forwarderFactoryMockCall struct {
	consumerGroup string
	outPublisher  message.Publisher
}

type forwarderFactoryMockReturn struct {
	forwarder *forwarder.Forwarder
	err       error
}

type forwarderFactoryMock struct {
	calls   []forwarderFactoryMockCall
	returns []forwarderFactoryMockReturn
}

func (f *forwarderFactoryMock) Forwarder(consumerGroup string, outPublisher message.Publisher) (*forwarder.Forwarder, error) {
	f.calls = append(f.calls, forwarderFactoryMockCall{
		consumerGroup: consumerGroup,
		outPublisher:  outPublisher,
	})
	if len(f.returns) == 0 {
		return &forwarder.Forwarder{}, nil
	}

	ret := f.returns[0]
	f.returns = f.returns[1:]

	return ret.forwarder, ret.err
}

func outboxFactoryCtxExtractorMock(publisher message.Publisher) func(ctx context.Context) *outboxFactoryMock {
	return func(ctx context.Context) *outboxFactoryMock {
		if publisher == nil {
			return nil
		}

		return &outboxFactoryMock{
			publisher: publisher,
		}
	}
}

type outboxFactoryMock struct {
	publisher message.Publisher
}

func (o *outboxFactoryMock) OutboxPublisher() (message.Publisher, error) {
	return o.publisher, nil
}

type publiserMockCall struct {
	topic string
	msgs  []*message.Message
}
type publisherMock struct {
	calls   []publiserMockCall
	returns []error
}

func (p *publisherMock) Publish(topic string, msgs ...*message.Message) error {
	p.calls = append(p.calls, publiserMockCall{
		topic: topic,
		msgs:  msgs,
	})
	if len(p.returns) == 0 {
		return nil
	}

	ret := p.returns[0]
	p.returns = p.returns[1:]

	return ret
}

func (p *publisherMock) Close() error {
	return nil
}

func TestPublish(t *testing.T) {
	t.Parallel()
	t.Run("publish success", testPublishSuccess)
	t.Run("publish error", testPublishError)
	t.Run("no factory in ctx", testNoFactoryInCtx)
}

func testPublishSuccess(t *testing.T) {
	pubmock := &publisherMock{}
	outboxPublisher := &rabbitmqOutboxPublisher[*outboxFactoryMock]{
		outPublisher: nil,
		msgForwarder: nil,

		outboxPublisherFactoryCtxExtractor: outboxFactoryCtxExtractorMock(pubmock),

		tracePropagator: tracing.NewWatermillTracePropagator(),
	}

	err := outboxPublisher.Publish(context.Background(), "topic", "test")
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}

	if len(pubmock.calls) != 1 {
		t.Errorf("expected 1 call, got %d", len(pubmock.calls))
	}

	if pubmock.calls[0].topic != "topic" {
		t.Errorf("expected topic to be topic, got %s", pubmock.calls[0].topic)
	}

	if len(pubmock.calls[0].msgs) != 1 {
		t.Errorf("expected 1 message, got %d", len(pubmock.calls[0].msgs))
	}

	if !slices.Equal(pubmock.calls[0].msgs[0].Payload, []byte(`"test"`)) {
		t.Errorf("expected payload to be %s, got %s", []byte(`"test"`), pubmock.calls[0].msgs[0].Payload)
	}
}

func testPublishError(t *testing.T) {
	pubmock := &publisherMock{
		returns: []error{eris.New("test")},
	}
	outboxPublisher := &rabbitmqOutboxPublisher[*outboxFactoryMock]{
		outPublisher: nil,
		msgForwarder: nil,

		outboxPublisherFactoryCtxExtractor: outboxFactoryCtxExtractorMock(pubmock),

		tracePropagator: tracing.NewWatermillTracePropagator(),
	}

	err := outboxPublisher.Publish(context.Background(), "topic", "test")
	if err == nil {
		t.Errorf("expected error, got nil")
	}

	if len(pubmock.calls) != 1 {
		t.Errorf("expected 1 call, got %d", len(pubmock.calls))
	}

	if pubmock.calls[0].topic != "topic" {
		t.Errorf("expected topic to be topic, got %s", pubmock.calls[0].topic)
	}

	if len(pubmock.calls[0].msgs) != 1 {
		t.Errorf("expected 1 message, got %d", len(pubmock.calls[0].msgs))
	}

	if !slices.Equal(pubmock.calls[0].msgs[0].Payload, []byte(`"test"`)) {
		t.Errorf("expected payload to be %s, got %s", []byte(`"test"`), pubmock.calls[0].msgs[0].Payload)
	}
}

func testNoFactoryInCtx(t *testing.T) {
	outboxPublisher := &rabbitmqOutboxPublisher[*outboxFactoryMock]{
		outPublisher: nil,
		msgForwarder: nil,

		outboxPublisherFactoryCtxExtractor: outboxFactoryCtxExtractorMock(nil),

		tracePropagator: tracing.NewWatermillTracePropagator(),
	}

	err := outboxPublisher.Publish(context.Background(), "topic", "test")
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}
