package manager

import (
	"time"

	"github.com/cenkalti/backoff/v4"
)

// ReconnectionBudget is how long we keep trying to get a working channel back
// once the transport was already up: a reconnection, or the re-registration
// that follows a successful one. When it is spent, the error is propagated and
// the process exits so that it can be restarted.
const ReconnectionBudget = 15 * time.Minute

// NewExponentialBackOff builds the exponential backoff shared by every amqp
// retry loop.
//
// A budget of zero or less means "no retries at all": backoff reads a zero
// MaxElapsedTime as "never stop", so we normalize it to a negative value, which
// makes the very first NextBackOff return Stop.
func NewExponentialBackOff(budget time.Duration) *backoff.ExponentialBackOff {
	if budget <= 0 {
		budget = -1
	}

	exponentialBackOff := backoff.NewExponentialBackOff()
	exponentialBackOff.MaxElapsedTime = budget

	return exponentialBackOff
}
