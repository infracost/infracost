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
	Unit  string
	Value any
}

type registeredMetric struct {
	Name    string
	Key     int
	Type    Type
	Context []string
	Metric  Metric
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

type Output []OutputMetric

type OutputMetric struct {
	Name   string        `json:"name"`
	Unit   string        `json:"unit,omitempty"`
	Values []OutputValue `json:"values,omitempty"`
}

type OutputValue struct {
	Context []string `json:"context,omitempty"`
	Value   any      `json:"value"`
}

func WriteMetrics(path string) error {
	coreRegistry.mu.RLock()
	defer coreRegistry.mu.RUnlock()

	outputMap := map[string]*OutputMetric{}

	for _, m := range coreRegistry.metrics {

		result := m.Metric.Result()

		metric := outputMap[m.Name]
		if metric == nil {
			metric = &OutputMetric{
				Name: m.Name,
				Unit: result.Unit,
			}
			outputMap[m.Name] = metric
		}

		metric.Values = append(metric.Values, OutputValue{
			Context: m.Context,
			Value:   result.Value,
		})
	}

	data, err := json.Marshal(outputMap)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}
