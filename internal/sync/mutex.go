package sync

import "sync"

// KeyMutex provides a concurrent safe way to acquire a lock based on an arbitrary string.
type KeyMutex struct {
	mutexes sync.Map // Zero value is empty and ready for use
}

// Lock locks a mutex for the given key.
func (m *KeyMutex) Lock(key string) func() {
	value, _ := m.mutexes.LoadOrStore(key, &sync.Mutex{})
	mtx := value.(*sync.Mutex)
	mtx.Lock()

	return func() { mtx.Unlock() }
}
