package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMetricsRecordsLowCardinalityHTTPMetrics(t *testing.T) {
	defaultMetrics = newMetricsStore()
	handler := Metrics("test-service")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/posts/post-123", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	metricsRec := httptest.NewRecorder()
	MetricsHandler("test-service").ServeHTTP(metricsRec, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	body := metricsRec.Body.String()

	// Check counter with status_group
	if !strings.Contains(body, `http_requests_total{service="test-service",method="POST",route="posts",status="201",status_group="2xx"} 1`) {
		t.Fatalf("missing request counter in metrics body:\n%s", body)
	}
	if strings.Contains(body, "post-123") {
		t.Fatalf("metrics body contains raw path id:\n%s", body)
	}
	if !strings.Contains(body, `http_requests_active{service="test-service"}`) {
		t.Fatalf("missing active request gauge:\n%s", body)
	}
	if !strings.Contains(body, "service_operations_total") {
		t.Fatalf("missing service operation counter:\n%s", body)
	}

	// Verify histogram buckets
	requiredBuckets := []string{"0.1", "0.3", "0.5", "1", "2", "5"}
	for _, le := range requiredBuckets {
		if !strings.Contains(body, fmt.Sprintf(`le="%s"`, le)) {
			t.Errorf("missing histogram bucket le=%s in metrics body:\n%s", le, body)
		}
	}

	// Verify status_group
	if !strings.Contains(body, `status_group="2xx"`) {
		t.Errorf("missing status_group=\"2xx\" label in metrics body:\n%s", body)
	}
}

func TestMetricsErrorGrouping(t *testing.T) {
	defaultMetrics = newMetricsStore()
	handler := Metrics("test-service")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/posts/123", nil)
	handler.ServeHTTP(httptest.NewRecorder(), req)

	metricsRec := httptest.NewRecorder()
	MetricsHandler("test-service").ServeHTTP(metricsRec, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	body := metricsRec.Body.String()

	if !strings.Contains(body, `status_group="5xx"`) {
		t.Errorf("expected status_group=\"5xx\" for 500 response, got:\n%s", body)
	}

	if !strings.Contains(body, `status="500"`) {
		t.Errorf("expected status=\"500\" in metrics body:\n%s", body)
	}
}
