package rabbitmq

import (
	"context"
	"io"
	"testing"

	"github.com/gothunder/thunder/pkg/events"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
)

// unreachableBroker points at a closed port on loopback, so the AMQP dial
// fails immediately instead of waiting on a network timeout.
const unreachableBroker = "amqp://guest:guest@127.0.0.1:1/"

type stubHandler struct{}

func (stubHandler) Topics() []string { return []string{"test.topic"} }

func (stubHandler) Handle(context.Context, string, events.EventDecoder) events.HandlerResponse {
	return events.Success
}

type stubNamedHandler struct{ stubHandler }

func (stubNamedHandler) QueuePosfix() string { return "test" }

func testLogger() *zerolog.Logger {
	logger := zerolog.New(io.Discard)

	return &logger
}

// A broker that is unreachable at startup used to be logged and ignored: the
// application started with no publisher and no consumer, reported healthy, and
// never retried. Startup must fail instead so the process can be restarted.
func TestUnreachableBrokerFailsStartup(t *testing.T) {
	tests := []struct {
		name   string
		option fx.Option
	}{
		{name: "publisher", option: PublisherModule},
		{name: "consumer", option: InvokeConsumer},
		{name: "named consumer", option: InvokeNamedConsumer},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("RABBIT_URL", unreachableBroker)

			app := fx.New(
				fx.NopLogger,
				fx.Supply(testLogger()),
				fx.Provide(func() events.Handler { return stubHandler{} }),
				fx.Provide(func() events.NamedHandler { return stubNamedHandler{} }),
				test.option,
			)

			require.Error(t, app.Err(), "an unreachable broker must abort startup")
		})
	}
}
