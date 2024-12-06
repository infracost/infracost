package metrics

import (
	"sync"
	"time"
)

type Timer struct {
	startTime   time.Time
	elapsedTime time.Duration
	running     bool
	mu          sync.Mutex
}

func (t *Timer) Result() Result {
	t.mu.Lock()
	defer t.mu.Unlock()
	elapsed := t.elapsedTime.Milliseconds()
	if t.running {
		elapsed += time.Since(t.startTime).Milliseconds()
	}
	return Result{
		Unit:  "ms",
		Value: elapsed,
	}
}

func GetTimer(name string, reuse bool, context ...string) *Timer {
	if reuse {
		existing := coreRegistry.Find(name, TypeTimer)
		if existing != nil {
			return existing.(*Timer)
		}
	}
	t := &Timer{}
	coreRegistry.Add(name, TypeTimer, t, context...)
	return t
}

func (t *Timer) Start() *Timer {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.running {
		return t
	}

	t.startTime = time.Now()
	t.running = true
	return t
}

func (t *Timer) Stop() *Timer {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.running {
		return t
	}

	t.elapsedTime += time.Since(t.startTime)
	t.running = false
	return t
}
