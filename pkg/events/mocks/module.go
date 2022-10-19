package mocks

import (
	context "context"

	"github.com/gothunder/thunder/internal/events/mocks"
	"github.com/gothunder/thunder/internal/events/mocks/consumer"
	"github.com/gothunder/thunder/internal/events/mocks/publisher"
	events "github.com/gothunder/thunder/pkg/events"
	"go.uber.org/fx"
)

func createMockChannel() chan mocks.MockedEvent {
	return make(chan mocks.MockedEvent)
}

func startConsumer(consumer events.EventConsumer, lc fx.Lifecycle, handler events.Handler) {
	lc.Append(
		fx.Hook{
			OnStart: func(ctx context.Context) error {
				go func() {
					consumer.Subscribe(ctx, handler)
				}()

				return nil
			},
			OnStop: func(ctx context.Context) error {
				consumer.Close(ctx)

				return nil
			},
		},
	)
}

var Module = fx.Options(
	fx.Provide(createMockChannel),
	fx.Provide(consumer.NewConsumer),
	fx.Provide(publisher.NewPublisher),
	fx.Invoke(startConsumer),
)
