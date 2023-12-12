package outbox

import (
	"context"
	"time"

	"github.com/TheRafaBonin/roxy"
	"github.com/ThreeDotsLabs/watermill/message"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

const (
	outboxRelayerMeterName = "thunder.outbox.relayer"
)

type withMetricsReayer struct {
	next                    Relayer
	relayerCounter          metric.Int64Counter
	relayerCounterErr       metric.Int64Counter
	relayerLatencyHistogram metric.Float64Histogram
}

// Close implements Relayer.
func (w *withMetricsReayer) Close(ctx context.Context) error {
	return w.next.Close(ctx)
}

// Start implements Relayer.
func (w *withMetricsReayer) Start(ctx context.Context) error {
	return w.next.Start(ctx)
}

// prepareMessages implements Relayer.
func (w *withMetricsReayer) prepareMessages(msgPack []Message) map[string][]*message.Message {
	return w.next.prepareMessages(msgPack)
}

// relay implements Relayer.
func (w *withMetricsReayer) relay(ctx context.Context, msgPack []Message) error {
	err := w.next.relay(ctx, msgPack)

	defer func() {
		topicCounter := make(map[string]int64)
		for _, msg := range msgPack {
			if _, ok := topicCounter[msg.Topic]; !ok {
				topicCounter[msg.Topic] = 0
			}
			topicCounter[msg.Topic]++
		}
		for topic, count := range topicCounter {
			w.relayerCounter.Add(ctx, count, metric.WithAttributes(
				semconv.MessagingDestinationName(topic),
			))
			if err != nil {
				w.relayerCounterErr.Add(ctx, count, metric.WithAttributes(
					semconv.MessagingDestinationName(topic),
				))
			}
		}

		for _, msg := range msgPack {
			w.relayerLatencyHistogram.Record(ctx, time.Since(msg.CreatedAt).Seconds(),
				metric.WithAttributes(
					semconv.MessagingDestinationName(msg.Topic),
				))
		}
	}()

	return err
}

func WrapRelayerWithMetrics(next Relayer) (Relayer, error) {
	meterProvider := otel.GetMeterProvider()

	meter := meterProvider.Meter(
		outboxRelayerMeterName,
	)

	relayerCounter, err := meter.Int64Counter("thunder.outbox.relayer.message.total",
		metric.WithDescription("Total number of messages relayed"),
		metric.WithUnit("1"))
	if err != nil {
		return nil, roxy.Wrap(err, "creating relayer total counter metrics")
	}

	relayerCounterErr, err := meter.Int64Counter("thunder.outbox.relayer.message.error",
		metric.WithDescription("Total number of messages relayed with error"),
		metric.WithUnit("1"))
	if err != nil {
		return nil, roxy.Wrap(err, "creating relayer error counter metrics")
	}

	relayerLatencyHistogram, err := meter.Float64Histogram("thunder.outbox.relayer.message.latency",
		metric.WithDescription("Latency of messages relayed in seconds"))
	if err != nil {
		return nil, roxy.Wrap(err, "creating relayer latency histogram metrics")
	}

	return &withMetricsReayer{
		next:                    next,
		relayerCounter:          relayerCounter,
		relayerCounterErr:       relayerCounterErr,
		relayerLatencyHistogram: relayerLatencyHistogram,
	}, nil
}
