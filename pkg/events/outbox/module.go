package outbox

import (
	"context"
	"errors"
	"reflect"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	internaloutbox "github.com/gothunder/thunder/internal/events/outbox"
	outboxent "github.com/gothunder/thunder/internal/events/outbox/ent"
	"github.com/gothunder/thunder/pkg/events/outbox/relayer"
	"github.com/gothunder/thunder/pkg/events/outbox/storer"
	"github.com/rs/zerolog"
	"go.uber.org/fx"
)

func provideStorer(s fx.Shutdowner, logger *zerolog.Logger) storer.Storer {
	str, err := storer.NewOutboxStorer(
		storer.WithTracing(),
		storer.WithMetrics(),
		storer.WithLogging())
	if err != nil {
		logger.Err(err).Msg("failed to create outbox storer")
		if err = s.Shutdown(); err != nil {
			logger.Err(err).Msg("failed to shutdown")
		}
	}

	return str
}

func resolveOutboxClient(entClient interface{}) interface{} {
	if entClient == nil {
		panic("entClient is nil")
	}
	oClient := reflect.ValueOf(entClient)
	if oClient.Kind() == reflect.Ptr {
		oClient = oClient.Elem()
	}
	oClient = oClient.FieldByName("OutboxMessage")
	if oClient.IsZero() {
		panic("no OutboxMessage in entClient")
	}

	return oClient.Interface()
}

func providePoller(pollInterval time.Duration, batchSize int) func(entClient interface{}, s fx.Shutdowner, logger *zerolog.Logger) relayer.MessagePoller {
	return func(entClient interface{}, s fx.Shutdowner, logger *zerolog.Logger) relayer.MessagePoller {
		oClient := resolveOutboxClient(entClient)

		poller, err := outboxent.NewEntMessagePoller(oClient, pollInterval, batchSize)
		if err != nil {
			logger.Err(err).Msg("failed to create outbox poller")
			if err = s.Shutdown(); err != nil {
				logger.Err(err).Msg("failed to shutdown")
			}
		}
		return poller
	}
}

func provideMarker(entClient interface{}, s fx.Shutdowner, logger *zerolog.Logger) relayer.MessageMarker {
	oClient := resolveOutboxClient(entClient)

	marker, err := outboxent.NewEntMessageMarker(oClient)
	if err != nil {
		logger.Err(err).Msg("failed to create outbox marker")
		if err = s.Shutdown(); err != nil {
			logger.Err(err).Msg("failed to shutdown")
		}
	}
	return marker
}

func startRelaying(
	lifecycle fx.Lifecycle,
	s fx.Shutdowner,
	logger *zerolog.Logger,
	entClient interface{},
	publisher message.Publisher,
	poller relayer.MessagePoller,
	marker relayer.MessageMarker,
) {
	r, err := relayer.NewOutboxRelayer(
		relayer.WithPoller(poller),
		relayer.WithPublisher(publisher),
		relayer.WithMarker(marker),
		relayer.WithTracing(),
		relayer.WithLogging(),
		relayer.WithMetrics(),
	)
	if err != nil {
		logger.Err(err).Msg("failed to create outbox relayer")
		if err = s.Shutdown(); err != nil {
			logger.Err(err).Msg("failed to shutdown")
		}
	}

	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			lctx := logger.WithContext(ctx)
			go func() {
				if err := r.Start(lctx); err != nil {
					if !errors.Is(err, context.Canceled) && !errors.Is(err, internaloutbox.ErrRelayerClosed) {
						logger.Err(err).Msg("failed relaying messages")
						if err = s.Shutdown(); err != nil {
							logger.Err(err).Msg("failed to shutdown")
						}
					}
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info().Msg("stopping relayer")

			// Create a new context with a timeout of 5 seconds
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			err = r.Close(ctx)
			if err != nil {
				logger.Err(err).Msg("failed to close relayer")
				return err
			}

			return nil
		},
	})
}

// it is meant to be used like following:
//
// fx.New(outbox.CreateModule(new(*ent.Client)))
func CreateModule(ppEntClient interface{}) fx.Option {
	pollInterval := 5 * time.Second
	batchSize := 100

	return fx.Options(
		fx.Provide(
			provideStorer,
			fx.Annotate(provideMarker,
				fx.From(
					ppEntClient,
					new(fx.Shutdowner),
					new(*zerolog.Logger))),
			fx.Annotate(providePoller(pollInterval, batchSize),
				fx.From(
					ppEntClient,
					new(fx.Shutdowner),
					new(*zerolog.Logger)))),
		fx.Invoke(
			fx.Annotate(
				startRelaying,
				fx.From(
					new(fx.Lifecycle),
					new(fx.Shutdowner),
					new(*zerolog.Logger),
					ppEntClient,
					new(message.Publisher),
					new(relayer.MessagePoller),
					new(relayer.MessageMarker))),
		),
	)
}
