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
	Type   Type          `json:"type"`
	Unit   string        `json:"unit"`
	Values []OutputValue `json:"values"`
}

type OutputValue struct {
	Context []string    `json:"context"`
	Value   interface{} `json:"value"`
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
				Type: m.Type,
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
	return os.WriteFile(path, data, 0644)
}
