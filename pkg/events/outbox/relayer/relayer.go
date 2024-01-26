package relayer

import (
	"sort"

	"github.com/TheRafaBonin/roxy"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/cenkalti/backoff/v4"
	internaloutbox "github.com/gothunder/thunder/internal/events/outbox"
)

type Relayer interface {
	internaloutbox.Relayer
}

type MessagePoller interface {
	internaloutbox.MessagePoller
}

type MessageMarker interface {
	internaloutbox.MessageMarker
}

func NewOutboxRelayer(options ...RelayerOptions) (Relayer, error) {
	cfg := newConfigDefaults()
	for _, opt := range options {
		opt(cfg)
	}
	sort.Slice(cfg.wrappers, func(i, j int) bool {
		return cfg.wrappers[i].order < cfg.wrappers[j].order
	})

	relayer, err := internaloutbox.NewRelayer(
		cfg.backOff,
		cfg.publisher,
		cfg.poller,
		cfg.marker,
	)
	if err != nil {
		return nil, roxy.Wrap(err, "creating relayer")
	}

	for _, wrapper := range cfg.wrappers {
		relayer, err = wrapper.wrap(relayer)
		if err != nil {
			return nil, roxy.Wrap(err, "wrapping relayer")
		}
	}
	return relayer, nil
}

type relayerConfig struct {
	publisher message.Publisher
	poller    MessagePoller
	marker    MessageMarker
	backOff   backoff.BackOff

	wrappers []relayerWrapper
}

type relayerWrapper struct {
	wrap  func(Relayer) (Relayer, error)
	order int // the higher the number, the more external the wrapper
}

type RelayerOptions func(*relayerConfig)

func newConfigDefaults() *relayerConfig {
	backOff := backoff.NewExponentialBackOff()
	backOff.MaxElapsedTime = 0
	return &relayerConfig{
		backOff:  backOff,
		wrappers: make([]relayerWrapper, 0),
	}
}

func WithBackOff(backOff backoff.BackOff) RelayerOptions {
	return func(cfg *relayerConfig) {
		cfg.backOff = backOff
	}
}

func WithPublisher(publisher message.Publisher) RelayerOptions {
	return func(cfg *relayerConfig) {
		cfg.publisher = publisher
	}
}

func WithPoller(poller MessagePoller) RelayerOptions {
	return func(cfg *relayerConfig) {
		cfg.poller = poller
	}
}

func WithMarker(marker MessageMarker) RelayerOptions {
	return func(cfg *relayerConfig) {
		cfg.marker = marker
	}
}

func WithTracing() RelayerOptions {
	return func(cfg *relayerConfig) {
		wrapper := relayerWrapper{
			wrap: func(relayer Relayer) (Relayer, error) {
				return internaloutbox.WrapRelayerWithTracing(relayer), nil
			},
			// tracing should be the most external wrapper
			// so it initializes the tracing span
			// and makes it acessible to the other wrappers
			order: 1,
		}
		cfg.wrappers = append(cfg.wrappers, wrapper)
	}
}

func WithLogging() RelayerOptions {
	return func(cfg *relayerConfig) {
		wrapper := relayerWrapper{
			wrap: func(relayer Relayer) (Relayer, error) {
				return internaloutbox.WrapRelayerWithLogging(relayer), nil
			},
			order: 0,
		}
		cfg.wrappers = append(cfg.wrappers, wrapper)
	}
}

func WithMetrics() RelayerOptions {
	return func(cfg *relayerConfig) {
		wrapper := relayerWrapper{
			wrap: func(relayer Relayer) (Relayer, error) {
				return internaloutbox.WrapRelayerWithMetrics(relayer)
			},
			order: 0,
		}
		cfg.wrappers = append(cfg.wrappers, wrapper)
	}
}
