package middleware

import (
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

	if !strings.Contains(body, `http_requests_total{service="test-service",method="POST",route="posts",status="201"} 1`) {
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
}
