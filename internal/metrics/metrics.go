package metrics

import (
	"encoding/json"
	"os"
	"sync"
)

type Type string

const (
	TypeCounter Type = "counter"
	TypeTimer   Type = "timer"
	TypeSetting Type = "setting"
)

type Metric interface {
	Result() Result
}

type Result struct {
	Unit  string      `json:"unit"`
	Value interface{} `json:"value"`
}

type registeredMetric struct {
	Name    string   `json:"name"`
	Key     int      `json:"key"`
	Type    Type     `json:"type"`
	Context []string `json:"context,omitempty"`
	Metric  Metric   `json:"-"`
}

type FullResult struct {
	registeredMetric
	Result
}

type registry struct {
	mu      sync.RWMutex
	metrics []registeredMetric
}

var coreRegistry = &registry{}

func (r *registry) Add(name string, t Type, m Metric, context ...string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	highestKey := -1
	for _, m := range r.metrics {
		if m.Name == name && m.Type == t && m.Key > highestKey {
			highestKey = m.Key
		}
	}
	r.metrics = append(r.metrics, registeredMetric{
		Name:    name,
		Type:    t,
		Metric:  m,
		Key:     highestKey + 1,
		Context: context,
	})
}

func (r *registry) Find(name string, t Type) Metric {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, m := range r.metrics {
		if m.Name == name && m.Type == t {
			return m.Metric
		}
	}
	return nil
}

func GetData() []FullResult {
	coreRegistry.mu.RLock()
	defer coreRegistry.mu.RUnlock()
	data := make([]FullResult, 0, len(coreRegistry.metrics))
	for _, m := range coreRegistry.metrics {
		data = append(data, FullResult{
			registeredMetric: m,
			Result:           m.Metric.Result(),
		})
	}
	return data
}

func WriteMetrics(path string) error {
	data, err := json.Marshal(GetData())
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
