package outbox

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/cenkalti/backoff/v4"
	"github.com/google/uuid"
)

func newBackoffMock(d time.Duration) backoff.BackOff {
	return backoff.NewConstantBackOff(d)
}

type messagePollerMock struct {
	msgPacks     [][]Message
	pollCalls    int
	closeResults []error
	closeCalls   int
}

func (m *messagePollerMock) Poll(ctx context.Context) (<-chan []Message, func(), error) {
	m.pollCalls++
	msgPacks := make(chan []Message, len(m.msgPacks))
	for _, msgPack := range m.msgPacks {
		msgPacks <- msgPack
	}
	close(msgPacks)
	return msgPacks, func() {}, nil
}

func (m *messagePollerMock) Close() error {
	m.closeCalls++
	if len(m.closeResults) == 0 {
		return nil
	}
	ret := m.closeResults[0]
	m.closeResults = m.closeResults[1:]
	return ret
}

func newMessagePullerMock(msgPacks [][]Message, closeResults []error) MessagePoller {
	return &messagePollerMock{
		msgPacks:     msgPacks,
		closeResults: closeResults,
	}
}

type messageMarkerMock struct {
	markResults []error
	markCalls   [][]Message
}

func (m *messageMarkerMock) MarkAsPublished(ctx context.Context, msgPack []Message) error {
	m.markCalls = append(m.markCalls, msgPack)
	if len(m.markResults) == 0 {
		return nil
	}
	ret := m.markResults[0]
	m.markResults = m.markResults[1:]
	return ret
}

func newMessageMarkerMock(markResults []error) MessageMarker {
	return &messageMarkerMock{
		markResults: markResults,
	}
}

type publisherMock struct {
	publishResults []error
	publisherCalls []struct {
		topic    string
		messages []*message.Message
	}
	closeResults []error
	closeCalls   int
	closed       bool
}

func (p *publisherMock) Publish(topic string, messages ...*message.Message) error {
	p.publisherCalls = append(p.publisherCalls, struct {
		topic    string
		messages []*message.Message
	}{
		topic:    topic,
		messages: messages,
	})
	if p.closed {
		return errors.New("closed")
	}
	if len(p.publishResults) == 0 {
		return nil
	}

	ret := p.publishResults[0]
	p.publishResults = p.publishResults[1:]
	return ret
}

func (p *publisherMock) Close() error {
	p.closeCalls++
	p.closed = true
	if len(p.closeResults) == 0 {
		return nil
	}
	ret := p.closeResults[0]
	p.closeResults = p.closeResults[1:]
	return ret
}

func newPublisherMock(publishResults []error) message.Publisher {
	return &publisherMock{
		publishResults: publishResults,
	}
}

func TestNewRelayer(t *testing.T) {
	tests := []struct {
		name        string
		backoff     backoff.BackOff
		publisher   message.Publisher
		poller      MessagePoller
		marker      MessageMarker
		expectedErr error
	}{
		{
			name:        "success",
			backoff:     nil,
			publisher:   newPublisherMock(nil),
			poller:      newMessagePullerMock(nil, nil),
			marker:      newMessageMarkerMock(nil),
			expectedErr: ErrNilBackoff,
		},
		{
			name:        "nil publisher",
			backoff:     newBackoffMock(0),
			publisher:   nil,
			poller:      newMessagePullerMock(nil, nil),
			marker:      newMessageMarkerMock(nil),
			expectedErr: ErrNilPublisher,
		},
		{
			name:        "nil poller",
			backoff:     newBackoffMock(0),
			publisher:   newPublisherMock(nil),
			poller:      nil,
			marker:      newMessageMarkerMock(nil),
			expectedErr: ErrNilPoller,
		},
		{
			name:        "nil marker",
			backoff:     newBackoffMock(0),
			publisher:   newPublisherMock(nil),
			poller:      newMessagePullerMock(nil, nil),
			marker:      nil,
			expectedErr: ErrNilMarker,
		},
		{
			name:        "success",
			backoff:     newBackoffMock(0),
			publisher:   newPublisherMock(nil),
			poller:      newMessagePullerMock(nil, nil),
			marker:      newMessageMarkerMock(nil),
			expectedErr: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			relayer, err := NewRelayer(test.backoff, test.publisher, test.poller, test.marker)
			if test.expectedErr != nil && !errors.Is(err, test.expectedErr) {
				t.Errorf("expected error %v, got %v", test.expectedErr, err)
			}
			if test.expectedErr != nil && relayer != nil {
				t.Errorf("expected nil relayer, got %v", relayer)
			}

			if test.expectedErr == nil && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
			if test.expectedErr == nil && relayer == nil {
				t.Errorf("expected non-nil relayer, got %v", relayer)
			}
		})
	}
}

func TestRelayer_Start(t *testing.T) {
	tests := []struct {
		name         string
		publisher    message.Publisher
		publishCalls []struct {
			topic    string
			messages []*message.Message
		}
		poller      MessagePoller
		marker      MessageMarker
		markerCalls [][]Message
		expectedErr error
	}{
		{
			name:      "success",
			publisher: newPublisherMock([]error{}),
			publishCalls: []struct {
				topic    string
				messages []*message.Message
			}{
				{
					topic: "topic",
					messages: []*message.Message{
						{
							UUID:    "8171478b-fece-4093-aa7e-342c4d816a21",
							Payload: []byte("payload"),
						},
					},
				},
			},
			poller: newMessagePullerMock([][]Message{
				{
					{
						ID:      uuid.MustParse("8171478b-fece-4093-aa7e-342c4d816a21"),
						Topic:   "topic",
						Payload: []byte("payload"),
					},
				},
			}, []error{}),
			marker: newMessageMarkerMock([]error{}),
			markerCalls: [][]Message{{
				{
					ID:      uuid.MustParse("8171478b-fece-4093-aa7e-342c4d816a21"),
					Topic:   "topic",
					Payload: []byte("payload"),
				},
			}},
			expectedErr: nil,
		},
		{
			name:      "publsher error once, then retry success",
			publisher: newPublisherMock([]error{errors.New("error")}),
			publishCalls: []struct {
				topic    string
				messages []*message.Message
			}{
				{
					topic: "topic",
					messages: []*message.Message{
						{
							UUID:    "8171478b-fece-4093-aa7e-342c4d816a21",
							Payload: []byte("payload"),
						},
					},
				},
				{
					topic: "topic",
					messages: []*message.Message{
						{
							UUID:    "8171478b-fece-4093-aa7e-342c4d816a21",
							Payload: []byte("payload"),
						},
					},
				},
			},
			poller: newMessagePullerMock([][]Message{
				{
					{
						ID:      uuid.MustParse("8171478b-fece-4093-aa7e-342c4d816a21"),
						Topic:   "topic",
						Payload: []byte("payload"),
					},
				},
			}, []error{}),
			marker: newMessageMarkerMock([]error{}),
			markerCalls: [][]Message{{
				{
					ID:      uuid.MustParse("8171478b-fece-4093-aa7e-342c4d816a21"),
					Topic:   "topic",
					Payload: []byte("payload"),
				},
			}},
			expectedErr: nil,
		},
		{
			name:      "marker error once, then retry success",
			publisher: newPublisherMock([]error{}),
			publishCalls: []struct {
				topic    string
				messages []*message.Message
			}{
				{
					topic: "topic",
					messages: []*message.Message{
						{
							UUID:    "8171478b-fece-4093-aa7e-342c4d816a21",
							Payload: []byte("payload"),
						},
					},
				},
				{
					topic: "topic",
					messages: []*message.Message{
						{
							UUID:    "8171478b-fece-4093-aa7e-342c4d816a21",
							Payload: []byte("payload"),
						},
					},
				},
			},
			poller: newMessagePullerMock([][]Message{
				{
					{
						ID:      uuid.MustParse("8171478b-fece-4093-aa7e-342c4d816a21"),
						Topic:   "topic",
						Payload: []byte("payload"),
					},
				},
			}, []error{}),
			marker: newMessageMarkerMock([]error{errors.New("error")}),
			markerCalls: [][]Message{{
				{
					ID:      uuid.MustParse("8171478b-fece-4093-aa7e-342c4d816a21"),
					Topic:   "topic",
					Payload: []byte("payload"),
				},
			}, {
				{
					ID:      uuid.MustParse("8171478b-fece-4093-aa7e-342c4d816a21"),
					Topic:   "topic",
					Payload: []byte("payload"),
				},
			}},
			expectedErr: nil,
		},
		{
			name:      "publisher error once, then marker error once, then retry success",
			publisher: newPublisherMock([]error{errors.New("error")}),
			publishCalls: []struct {
				topic    string
				messages []*message.Message
			}{
				{
					topic: "topic",
					messages: []*message.Message{
						{
							UUID:    "8171478b-fece-4093-aa7e-342c4d816a21",
							Payload: []byte("payload"),
						},
					},
				},
				{
					topic: "topic",
					messages: []*message.Message{
						{
							UUID:    "8171478b-fece-4093-aa7e-342c4d816a21",
							Payload: []byte("payload"),
						},
					},
				},
				{
					topic: "topic",
					messages: []*message.Message{
						{
							UUID:    "8171478b-fece-4093-aa7e-342c4d816a21",
							Payload: []byte("payload"),
						},
					},
				},
			},
			poller: newMessagePullerMock([][]Message{
				{
					{
						ID:      uuid.MustParse("8171478b-fece-4093-aa7e-342c4d816a21"),
						Topic:   "topic",
						Payload: []byte("payload"),
					},
				},
			}, []error{}),
			marker: newMessageMarkerMock([]error{errors.New("error")}),
			markerCalls: [][]Message{{
				{
					ID:      uuid.MustParse("8171478b-fece-4093-aa7e-342c4d816a21"),
					Topic:   "topic",
					Payload: []byte("payload"),
				},
			}, {
				{
					ID:      uuid.MustParse("8171478b-fece-4093-aa7e-342c4d816a21"),
					Topic:   "topic",
					Payload: []byte("payload"),
				},
			}},
			expectedErr: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			relayer, err := NewRelayer(newBackoffMock(0), test.publisher, test.poller, test.marker)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			relayer = WrapRelayerWithTracing(relayer)
			relayer, _ = WrapRelayerWithMetrics(relayer)
			relayer = WrapRelayerWithLogging(relayer)

			err = relayer.Start(context.Background())

			if test.expectedErr != err {
				t.Errorf("expected error %v, got %v", test.expectedErr, err)
			}

			if test.poller.(*messagePollerMock).pollCalls != 1 {
				t.Errorf("expected poller to be called once, got %d", test.poller.(*messagePollerMock).pollCalls)
			}

			if len(test.publisher.(*publisherMock).publisherCalls) != len(test.publishCalls) {
				t.Errorf("expected publisher to be called %d times, got %d", len(test.publishCalls), len(test.publisher.(*publisherMock).publisherCalls))
			}

			for i, call := range test.publishCalls {
				if test.publisher.(*publisherMock).publisherCalls[i].topic != call.topic {
					t.Errorf("expected publisher to be called with topic %s, got %s", call.topic, test.publisher.(*publisherMock).publisherCalls[i].topic)
				}
				if len(test.publisher.(*publisherMock).publisherCalls[i].messages) != len(call.messages) {
					t.Errorf("expected publisher to be called with %d messages, got %d", len(call.messages), len(test.publisher.(*publisherMock).publisherCalls[i].messages))
				}
				for j, msg := range call.messages {
					if test.publisher.(*publisherMock).publisherCalls[i].messages[j].UUID != msg.UUID {
						t.Errorf("expected publisher to be called with message %s, got %s", msg.UUID, test.publisher.(*publisherMock).publisherCalls[i].messages[j].UUID)
					}
					if string(test.publisher.(*publisherMock).publisherCalls[i].messages[j].Payload) != string(msg.Payload) {
						t.Errorf("expected publisher to be called with message payload %s, got %s", msg.Payload, test.publisher.(*publisherMock).publisherCalls[i].messages[j].Payload)
					}
				}
			}

			if len(test.marker.(*messageMarkerMock).markCalls) != len(test.markerCalls) {
				t.Errorf("expected marker to be called %d times, got %d", len(test.markerCalls), len(test.marker.(*messageMarkerMock).markCalls))
			}

			for i, call := range test.markerCalls {
				if len(test.marker.(*messageMarkerMock).markCalls[i]) != len(call) {
					t.Errorf("expected marker to be called with %d messages, got %d", len(call), len(test.marker.(*messageMarkerMock).markCalls[i]))
				}
				for j, msg := range call {
					if test.marker.(*messageMarkerMock).markCalls[i][j].ID != msg.ID {
						t.Errorf("expected marker to be called with message %s, got %s", msg.ID, test.marker.(*messageMarkerMock).markCalls[i][j].ID)
					}
					if test.marker.(*messageMarkerMock).markCalls[i][j].Topic != msg.Topic {
						t.Errorf("expected marker to be called with message topic %s, got %s", msg.Topic, test.marker.(*messageMarkerMock).markCalls[i][j].Topic)
					}
					if string(test.marker.(*messageMarkerMock).markCalls[i][j].Payload) != string(msg.Payload) {
						t.Errorf("expected marker to be called with message payload %s, got %s", msg.Payload, test.marker.(*messageMarkerMock).markCalls[i][j].Payload)
					}
				}
			}
		})
	}
}

func TestRelayer_Cancel(t *testing.T) {
	t.Run("Relayer cancel", func(t *testing.T) {
		publisher := newPublisherMock([]error{})
		poller := newMessagePullerMock([][]Message{
			{
				{
					ID:      uuid.MustParse("8171478b-fece-4093-aa7e-342c4d816a21"),
					Topic:   "topic",
					Payload: []byte("payload"),
				},
			},
		}, []error{})
		marker := newMessageMarkerMock([]error{})

		relayer, err := NewRelayer(
			newBackoffMock(0),
			publisher,
			poller,
			marker)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		relayer = WrapRelayerWithTracing(relayer)
		relayer, _ = WrapRelayerWithMetrics(relayer)
		relayer = WrapRelayerWithLogging(relayer)

		publisher.Close() // close publisher before starting relayer

		stop := make(chan struct{})

		go func() {
			err = relayer.Start(context.Background())
			close(stop)
		}()

		relayer.Close(context.Background())
		<-stop

		if !errors.Is(err, ErrRelayerClosed) {
			t.Errorf("expected ErrRelayerClosed, got %v", err)
		}
	})
}
