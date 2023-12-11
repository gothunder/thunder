package outbox

import (
	"context"

	"github.com/TheRafaBonin/roxy"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

const (
	outboxStorerMeterName = "thunder.outbox.storer"
)

type withMetricsStorer struct {
	next             Storer
	storedCounter    metric.Int64Counter
	storedCounterErr metric.Int64Counter
}

// Store implements Storer.
func (wts withMetricsStorer) Store(ctx context.Context, tx interface{}, messages []Message) (err error) {
	defer func() {
		topicCounter := make(map[string]int64)
		for _, msg := range messages {
			if _, ok := topicCounter[msg.Topic]; !ok {
				topicCounter[msg.Topic] = 0
			}
			topicCounter[msg.Topic]++
		}
		for topic, count := range topicCounter {
			wts.storedCounter.Add(ctx, count, metric.WithAttributes(
				semconv.MessagingDestinationName(topic),
			))
			if err != nil {
				wts.storedCounterErr.Add(ctx, int64(count), metric.WithAttributes(
					semconv.MessagingDestinationName(topic),
				))
			}
		}
	}()

	err = wts.next.Store(ctx, tx, messages)
	return
}

// WithTxClient implements Storer.
func (wts *withMetricsStorer) WithTxClient(tx interface{}) (TransactionalStorer, error) {
	// needs to be reimplemented here or otherwise TransactionalStorer
	// will not have metrics
	return newTransactionalStorer(wts, tx)
}

func WrapStorerWithMetrics(next Storer) (Storer, error) {
	meterProvider := otel.GetMeterProvider()

	meter := meterProvider.Meter(
		outboxStorerMeterName,
	)

	storedCounter, err := meter.Int64Counter(
		"thunder.outbox.storer.message.total",
		metric.WithDescription("Number of messages stored"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, roxy.Wrap(err, "creating stored counter")
	}

	storedCounterErr, err := meter.Int64Counter(
		"thunder.outbox.storer.message.error",
		metric.WithDescription("Number of messages stored with error"),
		metric.WithUnit("1"),
	)

	return &withMetricsStorer{
		next:             next,
		storedCounter:    storedCounter,
		storedCounterErr: storedCounterErr,
	}, nil
}
