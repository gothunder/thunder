package publisher

import "time"

func (r *rabbitmqPublisher) pause() {
	r.pausePublishMux.Lock()
	r.pausePublish = true
	r.pausePublishMux.Unlock()

	r.lastPausedAtMux.Lock()
	r.lastPausedAt = time.Now()
	r.lastPausedAtMux.Unlock()

	if len(r.pauseSignalChan) == 0 {
		r.pauseSignalChan <- true
	}
}

func (r *rabbitmqPublisher) resume() {
	r.pausePublishMux.Lock()
	r.pausePublish = false
	r.pausePublishMux.Unlock()
}

func (r *rabbitmqPublisher) isPaused() bool {
	r.pausePublishMux.RLock()
	defer r.pausePublishMux.RUnlock()
	return r.pausePublish
}

func (r *rabbitmqPublisher) getLastPausedAt() time.Time {
	r.lastPausedAtMux.RLock()
	defer r.lastPausedAtMux.RUnlock()
	return r.lastPausedAt
}
