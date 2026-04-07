package metrics

import "sync"

type Setting struct {
	value any
	mu    sync.Mutex
}

func (s *Setting) Result() Result {
	s.mu.Lock()
	defer s.mu.Unlock()
	return Result{
		Value: s.value,
	}
}

func GetSetting(name string, context ...string) *Setting {
	c := &Setting{}
	coreRegistry.Add(name, TypeSetting, c, context...)
	return c
}

func (s *Setting) Set(value any) *Setting {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.value = value
	return s
}
