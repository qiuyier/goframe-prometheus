package goframe_prometheus

import (
	"fmt"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"regexp"
	"time"
)

const namespace = "service"

var (
	labels = []string{"status", "endpoint", "method"}

	upTime = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "uptime",
			Help:      "HTTP service uptime.",
		}, nil,
	)

	reqCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "http_request_count_total",
			Help:      "Total number of HTTP requests made.",
		}, labels,
	)

	reqDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "http_request_duration_seconds",
			Help:      "HTTP request latencies in seconds.",
		}, labels,
	)

	reqSizeBytes = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: namespace,
			Name:      "http_request_size_bytes",
			Help:      "HTTP request sizes in bytes.",
		}, labels,
	)

	respSizeBytes = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: namespace,
			Name:      "http_response_size_bytes",
			Help:      "HTTP response sizes in bytes.",
		}, labels,
	)
)

type RequestLabelMappingFn func(c *ghttp.Request) string

// PromOpts represents the Prometheus middleware Options.
// It is used for filtering labels by regex.
type PromOpts struct {
	ExcludeRegexStatus     string
	ExcludeRegexEndpoint   string
	ExcludeRegexMethod     string
	EndpointLabelMappingFn RequestLabelMappingFn
}

func (po *PromOpts) checkLabel(label, pattern string) bool {
	if pattern == "" {
		return true
	}

	matched, err := regexp.MatchString(pattern, label)
	if err != nil {
		return true
	}
	return !matched
}

// calcRequestSize returns the size of request object.
func calcRequestSize(r *http.Request) float64 {
	size := 0
	if r.URL != nil {
		size = len(r.URL.String())
	}

	size += len(r.Method)
	size += len(r.Proto)

	for name, values := range r.Header {
		size += len(name)
		for _, value := range values {
			size += len(value)
		}
	}
	size += len(r.Host)

	// r.Form and r.MultipartForm are assumed to be included in r.URL.
	if r.ContentLength != -1 {
		size += int(r.ContentLength)
	}
	return float64(size)
}

// recordUptime increases service uptime per second.
func recordUptime() {
	for range time.Tick(time.Second) {
		upTime.WithLabelValues().Inc()
	}
}

// NewDefaultOpts return the default ProOpts
func NewDefaultOpts() *PromOpts {
	return &PromOpts{
		EndpointLabelMappingFn: func(c *ghttp.Request) string {
			//by default do nothing, return URL as is
			return c.Request.URL.Path
		},
	}
}

// PromMiddleWare a middleware for exporting some Web metrics
func PromMiddleWare(r *ghttp.Request) {
	promOpts := NewDefaultOpts()

	start := time.Now()
	r.Middleware.Next()

	status := fmt.Sprintf("%d", r.Response.Writer.Status)
	endpoint := promOpts.EndpointLabelMappingFn(r)
	method := r.Request.Method

	lvs := []string{status, endpoint, method}

	isOk := promOpts.checkLabel(status, promOpts.ExcludeRegexStatus) &&
		promOpts.checkLabel(endpoint, promOpts.ExcludeRegexEndpoint) &&
		promOpts.checkLabel(method, promOpts.ExcludeRegexMethod)

	if !isOk {
		return
	}

	respSize := len(r.GetBody())
	if respSize < 0 {
		respSize = 0
	}

	reqCount.WithLabelValues(lvs...).Inc()
	reqDuration.WithLabelValues(lvs...).Observe(time.Since(start).Seconds())
	reqSizeBytes.WithLabelValues(lvs...).Observe(calcRequestSize(r.Request))
	respSizeBytes.WithLabelValues(lvs...).Observe(float64(respSize))
}
