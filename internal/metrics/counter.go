package metrics

import "sync"

type Counter struct {
	value int
	mu    sync.Mutex
}

func (c *Counter) Result() Result {
	c.mu.Lock()
	defer c.mu.Unlock()
	return Result{
		Value: c.value,
	}
}

func GetCounter(name string, reuse bool, context ...string) *Counter {
	if reuse {
		existing := coreRegistry.Find(name, TypeCounter)
		if existing != nil {
			return existing.(*Counter)
		}
	}
	c := &Counter{}
	coreRegistry.Add(name, TypeCounter, c, context...)
	return c
}

func (c *Counter) Inc() *Counter {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value++
	return c
}

func (c *Counter) Add(value int) *Counter {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value += value
	return c
}
