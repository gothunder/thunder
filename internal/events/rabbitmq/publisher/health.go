package publisher

import (
	"sync"
	"sync/atomic"
	"time"
)

type waitGroupCount struct {
	sync.WaitGroup
	count int64
}

func (wg *waitGroupCount) Add(delta int) {
	atomic.AddInt64(&wg.count, int64(delta))
	wg.WaitGroup.Add(delta)
}

func (wg *waitGroupCount) Done() {
	atomic.AddInt64(&wg.count, -1)
	wg.WaitGroup.Done()
}

func (wg *waitGroupCount) GetCount() int64 {
	return atomic.LoadInt64(&wg.count)
}

func newWaitGroupCounter() *waitGroupCount {
	return &waitGroupCount{
		WaitGroup: sync.WaitGroup{},
		count:     0,
	}
}

const (
	timeWithoutPublishUnhealth = 30 * time.Second
	timePausedUnhealth         = 30 * time.Second
)

func (r *rabbitmqPublisher) healthCheckLoop() {
	logger := r.logger.With().Str("component", "publisher_health_check").Logger()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		<-ticker.C
		if r.closed.Load() {
			return
		}

		count := r.wg.GetCount()
		logger.Debug().Int64("messages_unpublished", count).Msg("checking publisher health")

		if r.isPaused() {
			logger.Debug().Time("last_paused", r.getLastPausedAt()).Msg("publisher is paused")
			if time.Since(r.getLastPausedAt()) > timePausedUnhealth {
				logger.Warn().
					Int64("messages_unpublished", count).
					Time("last_paused", r.getLastPausedAt()).
					Msg("publisher unhealthy: paused for too long")
			}
			continue
		}

		if count == 0 {
			logger.Debug().Msg("no messages pending, skipping")
			continue
		}

		r.lastPublishedAtMux.RLock()
		lastPublishedAt := r.lastPublishedAt
		r.lastPublishedAtMux.RUnlock()
		if time.Since(lastPublishedAt) > timeWithoutPublishUnhealth {
			logger.Warn().
				Int64("messages_unpublished", count).
				Time("last_published", lastPublishedAt).
				Msg("publisher unhealthy: no publishing for too long")
		}
	}
}
