package httpclient

import (
	"context"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Client wraps resty.Client with unified interface and telemetry support
type Client struct {
	client *resty.Client
	metrics         *HTTPMetrics
	tracer          trace.Tracer
	serviceName     string
	detailedMetrics bool
}

// Config holds HTTP client configuration
type Config struct {
	Timeout          time.Duration
	RetryCount       int
	RetryWaitTime    time.Duration
	RetryMaxWaitTime time.Duration
	ServiceName      string
	DetailedMetrics  bool
}

// DefaultConfig returns default HTTP client configuration
func DefaultConfig() Config {
	return Config{
		Timeout:          30 * time.Second,
		RetryCount:       3,
		RetryWaitTime:    1 * time.Second,
		RetryMaxWaitTime: 5 * time.Second,
		ServiceName:      "http-client",
		DetailedMetrics:  false,
	}
}

// NewClient creates a new HTTP client with unified interface and telemetry
func NewClient(cfg Config) *Client {
	restyClient := resty.New().
		SetTimeout(cfg.Timeout).
		SetRetryCount(cfg.RetryCount).
		SetRetryWaitTime(cfg.RetryWaitTime).
		SetRetryMaxWaitTime(cfg.RetryMaxWaitTime)

	client := &Client{
		client: restyClient,
	}

	
	// Initialize telemetry components
	serviceName := cfg.ServiceName
	if serviceName == "" {
		serviceName = "http-client"
	}

	client.metrics = NewHTTPMetrics(serviceName)
	client.tracer = otel.Tracer("http-client")
	client.serviceName = serviceName
	client.detailedMetrics = cfg.DetailedMetrics

	// Add OpenTelemetry instrumentation to transport
	restyClient.SetTransport(
		otelhttp.NewTransport(
			restyClient.GetClient().Transport,
			otelhttp.WithTracerProvider(otel.GetTracerProvider()),
		),
	)

	// Add hooks for metrics collection
	restyClient.OnBeforeRequest(client.beforeRequestHook)
	restyClient.OnAfterResponse(client.afterResponseHook)
	restyClient.OnError(client.errorHook)
	

	return client
}

// GetContext performs a GET request with context
func (c *Client) GetContext(ctx context.Context, url string) (*resty.Response, error) {
	
	return c.executeWithTelemetry(ctx, "GET", url, nil)
	
}

// PostContext performs a POST request with context
func (c *Client) PostContext(ctx context.Context, url string, body interface{}) (*resty.Response, error) {
	
	return c.executeWithTelemetry(ctx, "POST", url, body)
	
}

// PutContext performs a PUT request with context
func (c *Client) PutContext(ctx context.Context, url string, body interface{}) (*resty.Response, error) {
	
	return c.executeWithTelemetry(ctx, "PUT", url, body)
	
}

// PatchContext performs a PATCH request with context
func (c *Client) PatchContext(ctx context.Context, url string, body interface{}) (*resty.Response, error) {
	
	return c.executeWithTelemetry(ctx, "PATCH", url, body)
	
}

// DeleteContext performs a DELETE request with context
func (c *Client) DeleteContext(ctx context.Context, url string) (*resty.Response, error) {
	
	return c.executeWithTelemetry(ctx, "DELETE", url, nil)
	
}


// executeWithTelemetry executes an HTTP request with full telemetry
func (c *Client) executeWithTelemetry(ctx context.Context, method, url string, body interface{}) (*resty.Response, error) {
	start := time.Now()

	// Create custom span
	ctx, span := c.tracer.Start(ctx, fmt.Sprintf("http.%s", method),
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()

	// Add span attributes
	span.SetAttributes(
		attribute.String("http.method", method),
		attribute.String("http.url", c.sanitizeURL(url)),
	)

	// Build request
	req := c.client.R().SetContext(ctx)

	if body != nil {
		req.SetBody(body)
		
	}

	// Execute request based on method
	var resp *resty.Response
	var err error

	switch method {
	case "GET":
		resp, err = req.Get(url)
	case "POST":
		resp, err = req.Post(url)
	case "PUT":
		resp, err = req.Put(url)
	case "PATCH":
		resp, err = req.Patch(url)
	case "DELETE":
		resp, err = req.Delete(url)
	default:
		err = fmt.Errorf("unsupported HTTP method: %s", method)
	}

	duration := time.Since(start).Seconds()

	// Record metrics and span status
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		c.metrics.RecordRequest(method, url, 0, "error", duration)
		c.metrics.RecordError(method, categorizeHTTPError(err))
		return resp, err
	}

	statusCode := resp.StatusCode()
	status := getStatusCategory(statusCode)

	span.SetAttributes(
		attribute.Int("http.status_code", statusCode),
		attribute.Int64("http.response.body.size", int64(len(resp.Body()))),
	)

	if statusCode >= 400 {
		span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", statusCode))
	} else {
		span.SetStatus(codes.Ok, "")
	}

	c.metrics.RecordRequest(method, url, statusCode, status, duration)
	

	return resp, nil
}

// beforeRequestHook is called before each request
func (c *Client) beforeRequestHook(client *resty.Client, req *resty.Request) error {
	ctx := req.Context()
	ctx = context.WithValue(ctx, "request_start_time", time.Now())
	req.SetContext(ctx)
	return nil
}

// afterResponseHook is called after each successful response
func (c *Client) afterResponseHook(client *resty.Client, resp *resty.Response) error {
	if resp.Request.Attempt > 1 {
		c.metrics.RecordRetry(resp.Request.Method, resp.Request.URL)
	}
	return nil
}

// errorHook is called when request fails
func (c *Client) errorHook(req *resty.Request, err error) {
	if req.Attempt > 1 {
		c.metrics.RecordRetry(req.Method, req.URL)
	}
}

// sanitizeURL removes sensitive information from URL for tracing
func (c *Client) sanitizeURL(url string) string {
	if !c.detailedMetrics {
		return "***"
	}
	return url
}

// getStatusCategory categorizes HTTP status codes
func getStatusCategory(statusCode int) string {
	switch {
	case statusCode >= 200 && statusCode < 300:
		return "2xx"
	case statusCode >= 300 && statusCode < 400:
		return "3xx"
	case statusCode >= 400 && statusCode < 500:
		return "4xx"
	case statusCode >= 500:
		return "5xx"
	default:
		return "unknown"
	}
}

// categorizeHTTPError categorizes HTTP errors for metrics
func categorizeHTTPError(err error) string {
	if err == nil {
		return "none"
	}

	errStr := err.Error()
	switch {
	case contains(errStr, "timeout"):
		return "timeout"
	case contains(errStr, "connection refused"):
		return "connection_refused"
	case contains(errStr, "connection reset"):
		return "connection_reset"
	case contains(errStr, "no such host"):
		return "dns_error"
	case contains(errStr, "TLS"):
		return "tls_error"
	case contains(errStr, "context canceled"):
		return "canceled"
	case contains(errStr, "context deadline exceeded"):
		return "deadline_exceeded"
	default:
		return "unknown_error"
	}
}

// estimateBodySize estimates the size of request body
func estimateBodySize(body interface{}) int64 {
	if body == nil {
		return 0
	}

	switch v := body.(type) {
	case string:
		return int64(len(v))
	case []byte:
		return int64(len(v))
	default:
		return 0
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr))
}

