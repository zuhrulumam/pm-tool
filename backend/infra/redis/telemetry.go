package redis

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// RedisMetrics holds Prometheus metrics for Redis operations
type RedisMetrics struct {
	// Operation metrics
	operationsTotal   *prometheus.CounterVec
	operationDuration *prometheus.HistogramVec
	
	// Cache metrics
	cacheHits *prometheus.CounterVec
	
	// Connection pool metrics
	connectionPoolHits     prometheus.Counter
	connectionPoolMisses   prometheus.Counter
	connectionPoolTimeouts prometheus.Counter
	connectionPoolSize     prometheus.Gauge
	connectionPoolIdle     prometheus.Gauge
	connectionPoolStale    prometheus.Gauge
	
	// Error metrics
	errorsTotal *prometheus.CounterVec
	
	// Key size metrics
	keySize *prometheus.HistogramVec
	
	// Connection state
	connected prometheus.Gauge
	
	mu sync.RWMutex
}

// NewRedisMetrics creates a new Redis metrics collector
func NewRedisMetrics(serviceName string) *RedisMetrics {
	metrics := &RedisMetrics{
		operationsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "redis_operations_total",
				Help: "Total number of Redis operations",
				ConstLabels: prometheus.Labels{
					"service": serviceName,
				},
			},
			[]string{"command", "status"},
		),
		operationDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "redis_operation_duration_seconds",
				Help:    "Duration of Redis operations in seconds",
				Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5},
				ConstLabels: prometheus.Labels{
					"service": serviceName,
				},
			},
			[]string{"command", "status"},
		),
		cacheHits: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "redis_cache_hits_total",
				Help: "Total number of cache hits and misses",
				ConstLabels: prometheus.Labels{
					"service": serviceName,
				},
			},
			[]string{"command", "result"},
		),
		connectionPoolHits: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "redis_connection_pool_hits_total",
				Help: "Total number of connection pool hits",
				ConstLabels: prometheus.Labels{
					"service": serviceName,
				},
			},
		),
		connectionPoolMisses: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "redis_connection_pool_misses_total",
				Help: "Total number of connection pool misses",
				ConstLabels: prometheus.Labels{
					"service": serviceName,
				},
			},
		),
		connectionPoolTimeouts: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "redis_connection_pool_timeouts_total",
				Help: "Total number of connection pool timeouts",
				ConstLabels: prometheus.Labels{
					"service": serviceName,
				},
			},
		),
		connectionPoolSize: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "redis_connection_pool_size",
				Help: "Current size of the connection pool",
				ConstLabels: prometheus.Labels{
					"service": serviceName,
				},
			},
		),
		connectionPoolIdle: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "redis_connection_pool_idle",
				Help: "Number of idle connections in the pool",
				ConstLabels: prometheus.Labels{
					"service": serviceName,
				},
			},
		),
		connectionPoolStale: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "redis_connection_pool_stale",
				Help: "Number of stale connections in the pool",
				ConstLabels: prometheus.Labels{
					"service": serviceName,
				},
			},
		),
		errorsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "redis_errors_total",
				Help: "Total number of Redis errors by type",
				ConstLabels: prometheus.Labels{
					"service": serviceName,
				},
			},
			[]string{"command", "error_type"},
		),
		keySize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "redis_key_size_bytes",
				Help:    "Size of Redis keys/values in bytes",
				Buckets: prometheus.ExponentialBuckets(64, 2, 12), // 64B to 256KB
				ConstLabels: prometheus.Labels{
					"service": serviceName,
				},
			},
			[]string{"command"},
		),
		connected: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "redis_connected",
				Help: "Whether Redis client is connected (1) or not (0)",
				ConstLabels: prometheus.Labels{
					"service": serviceName,
				},
			},
		),
	}
	
	return metrics
}

// RecordOperation records a Redis operation with duration and status
func (m *RedisMetrics) RecordOperation(command, status string, duration float64) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	m.operationsTotal.WithLabelValues(command, status).Inc()
	m.operationDuration.WithLabelValues(command, status).Observe(duration)
}

// RecordCacheHit records a cache hit or miss
func (m *RedisMetrics) RecordCacheHit(command string, hit bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	result := "miss"
	if hit {
		result = "hit"
	}
	m.cacheHits.WithLabelValues(command, result).Inc()
}

// RecordError records a Redis error by type
func (m *RedisMetrics) RecordError(command, errorType string) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	m.errorsTotal.WithLabelValues(command, errorType).Inc()
}

// RecordKeySize records the size of a key or value
func (m *RedisMetrics) RecordKeySize(command string, size int64) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	m.keySize.WithLabelValues(command).Observe(float64(size))
}

// RecordConnectionPool records connection pool statistics
func (m *RedisMetrics) RecordConnectionPool(hits, misses, timeouts, total, idle, stale uint32) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	m.connectionPoolHits.Add(float64(hits))
	m.connectionPoolMisses.Add(float64(misses))
	m.connectionPoolTimeouts.Add(float64(timeouts))
	m.connectionPoolSize.Set(float64(total))
	m.connectionPoolIdle.Set(float64(idle))
	m.connectionPoolStale.Set(float64(stale))
}

// RecordConnection records connection state
func (m *RedisMetrics) RecordConnection(connected bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if connected {
		m.connected.Set(1)
	} else {
		m.connected.Set(0)
	}
}

// Reset resets all metrics (useful for testing)
func (m *RedisMetrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.operationsTotal.Reset()
	m.operationDuration.Reset()
	m.cacheHits.Reset()
	m.errorsTotal.Reset()
	m.keySize.Reset()
}
