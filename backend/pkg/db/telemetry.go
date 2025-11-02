package db

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// DBMetrics holds Prometheus metrics for database operations
type DBMetrics struct {
	queriesTotal   *prometheus.CounterVec
	queryDuration  *prometheus.HistogramVec
	errorsTotal    *prometheus.CounterVec
	connections    *prometheus.GaugeVec
	mu             sync.RWMutex
}

// NewDBMetrics creates a new database metrics collector
func NewDBMetrics(serviceName string) *DBMetrics {
	return &DBMetrics{
		queriesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "db_queries_total",
				Help: "Total number of database queries",
				ConstLabels: prometheus.Labels{
					"service": serviceName,
				},
			},
			[]string{"operation", "status"},
		),
		queryDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "db_query_duration_seconds",
				Help:    "Duration of database queries in seconds",
				Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
				ConstLabels: prometheus.Labels{
					"service": serviceName,
				},
			},
			[]string{"operation", "status"},
		),
		errorsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "db_errors_total",
				Help: "Total number of database errors by type",
				ConstLabels: prometheus.Labels{
					"service": serviceName,
				},
			},
			[]string{"operation", "error_type"},
		),
		connections: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "db_connections",
				Help: "Number of database connections",
				ConstLabels: prometheus.Labels{
					"service": serviceName,
				},
			},
			[]string{"state"},
		),
	}
}

// RecordQuery records a database query with duration and status
func (m *DBMetrics) RecordQuery(operation, status string, duration float64) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	m.queriesTotal.WithLabelValues(operation, status).Inc()
	m.queryDuration.WithLabelValues(operation, status).Observe(duration)
}

// RecordError records a database error by type
func (m *DBMetrics) RecordError(operation, errorType string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	m.errorsTotal.WithLabelValues(operation, errorType).Inc()
}

// SetConnections sets the number of connections
func (m *DBMetrics) SetConnections(open, idle, inUse int) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	m.connections.WithLabelValues("open").Set(float64(open))
	m.connections.WithLabelValues("idle").Set(float64(idle))
	m.connections.WithLabelValues("in_use").Set(float64(inUse))
}

// Reset resets all metrics (useful for testing)
func (m *DBMetrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.queriesTotal.Reset()
	m.queryDuration.Reset()
	m.errorsTotal.Reset()
}
