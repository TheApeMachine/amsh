package system

import (
	"sync"
	"time"
)

type MetricType string

const (
	MetricTypeProcessDuration MetricType = "process_duration"
	MetricTypeAgentResponse   MetricType = "agent_response"
	MetricTypeToolUsage       MetricType = "tool_usage"
)

type Metric struct {
	Type      MetricType
	Value     float64
	Timestamp time.Time
	Labels    map[string]string
}

type Monitor struct {
	metrics []Metric
	mu      sync.RWMutex
}

/*
NewMonitor creates a new monitor instance.
*/
func NewMonitor() *Monitor {
	return &Monitor{
		metrics: make([]Metric, 0),
	}
}

func (m *Monitor) RecordMetric(metricType MetricType, value float64, labels map[string]string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.metrics = append(m.metrics, Metric{
		Type:      metricType,
		Value:     value,
		Timestamp: time.Now(),
		Labels:    labels,
	})
}

func (m *Monitor) Report() []Metric {
	m.mu.RLock()
	defer m.mu.RUnlock()

	metrics := make([]Metric, len(m.metrics))
	copy(metrics, m.metrics)
	return metrics
}
