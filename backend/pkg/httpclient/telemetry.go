package httpclient

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// HTTPMetrics holds Prometheus metrics for HTTP operations
type HTTPMetrics struct {
	requestsTotal   *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
	retriesTotal    *prometheus.CounterVec
	errorsTotal     *prometheus.CounterVec
	responseSize    *prometheus.HistogramVec
	requestSize     *prometheus.HistogramVec
	activeRequests  *prometheus.GaugeVec
	mu              sync.RWMutex
}

// NewHTTPMetrics creates a new HTTP metrics collector
func NewHTTPMetrics(serviceName string) *HTTPMetrics {
	return &HTTPMetrics{
		requestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_client_requests_total",
				Help: "Total number of HTTP requests",
				ConstLabels: prometheus.Labels{
					"service": serviceName,
				},
			},
			[]string{"method", "endpoint", "status"},
		),
		requestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_client_request_duration_seconds",
				Help:    "Duration of HTTP requests in seconds",
				Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
				ConstLabels: prometheus.Labels{
					"service": serviceName,
				},
			},
			[]string{"method", "endpoint", "status"},
		),
		retriesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_client_retries_total",
				Help: "Total number of HTTP request retries",
				ConstLabels: prometheus.Labels{
					"service": serviceName,
				},
			},
			[]string{"method", "endpoint"},
		),
		errorsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_client_errors_total",
				Help: "Total number of HTTP errors by type",
				ConstLabels: prometheus.Labels{
					"service": serviceName,
				},
			},
			[]string{"method", "error_type"},
		),
		responseSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_client_response_size_bytes",
				Help:    "Size of HTTP response bodies in bytes",
				Buckets: prometheus.ExponentialBuckets(100, 10, 8),
				ConstLabels: prometheus.Labels{
					"service": serviceName,
				},
			},
			[]string{"method"},
		),
		requestSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_client_request_size_bytes",
				Help:    "Size of HTTP request bodies in bytes",
				Buckets: prometheus.ExponentialBuckets(100, 10, 8),
				ConstLabels: prometheus.Labels{
					"service": serviceName,
				},
			},
			[]string{"method"},
		),
		activeRequests: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "http_client_active_requests",
				Help: "Number of active HTTP requests",
				ConstLabels: prometheus.Labels{
					"service": serviceName,
				},
			},
			[]string{"method", "endpoint"},
		),
	}
}

// RecordRequest records an HTTP request with duration and status
func (m *HTTPMetrics) RecordRequest(method, endpoint string, statusCode int, status string, duration float64) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sanitizedEndpoint := sanitizeEndpoint(endpoint)
	m.requestsTotal.WithLabelValues(method, sanitizedEndpoint, status).Inc()
	m.requestDuration.WithLabelValues(method, sanitizedEndpoint, status).Observe(duration)
}

// RecordRetry records an HTTP request retry
func (m *HTTPMetrics) RecordRetry(method, endpoint string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sanitizedEndpoint := sanitizeEndpoint(endpoint)
	m.retriesTotal.WithLabelValues(method, sanitizedEndpoint).Inc()
}

// RecordError records an HTTP error by type
func (m *HTTPMetrics) RecordError(method, errorType string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	m.errorsTotal.WithLabelValues(method, errorType).Inc()
}

// RecordResponseSize records the size of an HTTP response
func (m *HTTPMetrics) RecordResponseSize(method string, size int64) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	m.responseSize.WithLabelValues(method).Observe(float64(size))
}

// RecordRequestSize records the size of an HTTP request
func (m *HTTPMetrics) RecordRequestSize(method string, size int64) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	m.requestSize.WithLabelValues(method).Observe(float64(size))
}

// IncActiveRequests increments the active requests counter
func (m *HTTPMetrics) IncActiveRequests(method, endpoint string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sanitizedEndpoint := sanitizeEndpoint(endpoint)
	m.activeRequests.WithLabelValues(method, sanitizedEndpoint).Inc()
}

// DecActiveRequests decrements the active requests counter
func (m *HTTPMetrics) DecActiveRequests(method, endpoint string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sanitizedEndpoint := sanitizeEndpoint(endpoint)
	m.activeRequests.WithLabelValues(method, sanitizedEndpoint).Dec()
}

// Reset resets all metrics (useful for testing)
func (m *HTTPMetrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.requestsTotal.Reset()
	m.requestDuration.Reset()
	m.retriesTotal.Reset()
	m.errorsTotal.Reset()
	m.responseSize.Reset()
	m.requestSize.Reset()
}

// sanitizeEndpoint extracts the path pattern from URL to avoid high cardinality
func sanitizeEndpoint(endpoint string) string {
	// Simple implementation - return "external" for all external calls
	// In production, parse the URL and extract the path pattern
	return "external"
}
