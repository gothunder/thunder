package manager

import (
	"io"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rotisserie/eris"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

// The first dial is the only one with no reconnection loop behind it, so it has
// to retry on its own: a broker that is restarting used to kill every pod that
// happened to boot during the window.
func TestConnectWithRetry(t *testing.T) {
	logger := zerolog.New(io.Discard)

	failingDialer := func(attempts *int) dialer {
		return func() (*amqp.Connection, *amqp.Channel, error) {
			*attempts++
			return nil, nil, eris.New("broker is down")
		}
	}

	t.Run("a budget of zero dials exactly once", func(t *testing.T) {
		attempts := 0

		_, _, err := connectWithRetry(failingDialer(&attempts), &logger, 0)

		require.Error(t, err, "an unreachable broker must still report the failure")
		require.Equal(t, 1, attempts, "a zero budget means no retries")
	})

	t.Run("a budget retries and then gives up", func(t *testing.T) {
		attempts := 0

		started := time.Now()
		_, _, err := connectWithRetry(failingDialer(&attempts), &logger, time.Second)

		require.Error(t, err, "the error must be propagated once the budget is spent")
		require.Greater(t, attempts, 1, "the first dial must be retried")
		require.Less(t, time.Since(started), 5*time.Second, "the budget must bound the wait")
	})

	t.Run("it stops at the first successful dial", func(t *testing.T) {
		attempts := 0

		_, _, err := connectWithRetry(func() (*amqp.Connection, *amqp.Channel, error) {
			attempts++
			if attempts == 1 {
				return nil, nil, eris.New("broker is down")
			}

			return nil, nil, nil
		}, &logger, time.Second)

		require.NoError(t, err)
		require.Equal(t, 2, attempts)
	})
}
