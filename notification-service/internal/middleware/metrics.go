package middleware

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var latencyBuckets = []float64{0.1, 0.3, 0.5, 1, 2, 5}

type metricKey struct {
	Service     string
	Method      string
	Route       string
	Status      string
	StatusGroup string
}

type metricSnapshot struct {
	Requests   map[metricKey]uint64
	Buckets    map[metricKey][]uint64
	Duration   map[metricKey]float64
	Operations map[metricKey]uint64
	Active     int64
}

type metricsStore struct {
	mu         sync.Mutex
	requests   map[metricKey]uint64
	buckets    map[metricKey][]uint64
	duration   map[metricKey]float64
	operations map[metricKey]uint64
	active     int64
}

func newMetricsStore() *metricsStore {
	return &metricsStore{
		requests:   make(map[metricKey]uint64),
		buckets:    make(map[metricKey][]uint64),
		duration:   make(map[metricKey]float64),
		operations: make(map[metricKey]uint64),
	}
}

var defaultMetrics = newMetricsStore()

func Metrics(serviceName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defaultMetrics.incActive()
			started := time.Now()
			rw := &metricsResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			defer func() {
				defaultMetrics.decActive()
				defaultMetrics.observe(serviceName, r.Method, metricRoute(r.URL.Path), rw.statusCode, time.Since(started).Seconds())
			}()

			next.ServeHTTP(rw, r)
		})
	}
}

func MetricsHandler(serviceName string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		_, _ = w.Write([]byte(defaultMetrics.render(serviceName)))
	})
}

type metricsResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *metricsResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (s *metricsStore) incActive() {
	s.mu.Lock()
	s.active++
	s.mu.Unlock()
}

func (s *metricsStore) decActive() {
	s.mu.Lock()
	s.active--
	s.mu.Unlock()
}

func statusGroup(code int) string {
	switch {
	case code >= 200 && code < 300:
		return "2xx"
	case code >= 400 && code < 500:
		return "4xx"
	case code >= 500:
		return "5xx"
	default:
		return "other"
	}
}

func (s *metricsStore) observe(serviceName string, method string, route string, status int, seconds float64) {
	key := metricKey{
		Service:     serviceName,
		Method:      method,
		Route:       route,
		Status:      strconv.Itoa(status),
		StatusGroup: statusGroup(status),
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	s.requests[key]++
	s.duration[key] += seconds
	s.operations[key]++
	if _, ok := s.buckets[key]; !ok {
		s.buckets[key] = make([]uint64, len(latencyBuckets))
	}
	for i, bucket := range latencyBuckets {
		if seconds <= bucket {
			s.buckets[key][i]++
		}
	}
}

func (s *metricsStore) snapshot() metricSnapshot {
	s.mu.Lock()
	defer s.mu.Unlock()

	snap := metricSnapshot{
		Requests:   make(map[metricKey]uint64, len(s.requests)),
		Buckets:    make(map[metricKey][]uint64, len(s.buckets)),
		Duration:   make(map[metricKey]float64, len(s.duration)),
		Operations: make(map[metricKey]uint64, len(s.operations)),
		Active:     s.active,
	}
	for k, v := range s.requests {
		snap.Requests[k] = v
	}
	for k, v := range s.buckets {
		cp := make([]uint64, len(v))
		copy(cp, v)
		snap.Buckets[k] = cp
	}
	for k, v := range s.duration {
		snap.Duration[k] = v
	}
	for k, v := range s.operations {
		snap.Operations[k] = v
	}
	return snap
}

func (s *metricsStore) render(serviceName string) string {
	snap := s.snapshot()
	keys := sortedMetricKeys(snap.Requests)
	var b strings.Builder

	b.WriteString("# HELP http_requests_total Total HTTP requests.\n")
	b.WriteString("# TYPE http_requests_total counter\n")
	for _, key := range keys {
		fmt.Fprintf(&b, "http_requests_total{%s} %d\n", key.labels(), snap.Requests[key])
	}

	b.WriteString("# HELP http_request_duration_seconds HTTP request duration histogram.\n")
	b.WriteString("# TYPE http_request_duration_seconds histogram\n")
	for _, key := range keys {
		for i, bucket := range latencyBuckets {
			fmt.Fprintf(&b, "http_request_duration_seconds_bucket{%s,le=%q} %d\n", key.labels(), formatFloat(bucket), snap.Buckets[key][i])
		}
		fmt.Fprintf(&b, "http_request_duration_seconds_bucket{%s,le=\"+Inf\"} %d\n", key.labels(), snap.Requests[key])
		fmt.Fprintf(&b, "http_request_duration_seconds_sum{%s} %s\n", key.labels(), formatFloat(snap.Duration[key]))
		fmt.Fprintf(&b, "http_request_duration_seconds_count{%s} %d\n", key.labels(), snap.Requests[key])
	}

	b.WriteString("# HELP http_requests_active Active HTTP requests.\n")
	b.WriteString("# TYPE http_requests_active gauge\n")
	fmt.Fprintf(&b, "http_requests_active{service=%q} %d\n", serviceName, snap.Active)

	b.WriteString("# HELP service_operations_total Total service operations grouped by route and status.\n")
	b.WriteString("# TYPE service_operations_total counter\n")
	for _, key := range sortedMetricKeys(snap.Operations) {
		fmt.Fprintf(&b, "service_operations_total{%s} %d\n", key.labels(), snap.Operations[key])
	}

	return b.String()
}

func sortedMetricKeys(values map[metricKey]uint64) []metricKey {
	keys := make([]metricKey, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].labels() < keys[j].labels()
	})
	return keys
}

func (k metricKey) labels() string {
	return fmt.Sprintf(
		"service=%q,method=%q,route=%q,status=%q,status_group=%q",
		k.Service, k.Method, k.Route, k.Status, k.StatusGroup,
	)
}

func formatFloat(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func metricRoute(path string) string {
	switch {
	case path == "/health":
		return "health"
	case strings.HasPrefix(path, "/metrics"):
		return "metrics"
	case strings.HasPrefix(path, "/api/v1/auth"):
		return "auth"
	case strings.HasPrefix(path, "/api/v1/users"):
		return "users"
	case strings.HasPrefix(path, "/api/v1/posts"):
		return "posts"
	case strings.HasPrefix(path, "/api/v1/feed"):
		return "feed"
	case strings.HasPrefix(path, "/api/v1/notifications"):
		return "notifications"
	default:
		return "unknown"
	}
}
