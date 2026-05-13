package middleware

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func TestMetricsRecordsLowCardinalityHTTPMetrics(t *testing.T) {
	metricsServer := httptest.NewServer(promhttp.Handler())
	defer metricsServer.Close()

	handler := Metrics("test-service")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/posts/post-123", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	resp, err := http.Get(metricsServer.URL)
	if err != nil {
		t.Fatalf("failed to fetch metrics: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read metrics body: %v", err)
	}
	bodyStr := string(bodyBytes)

	// Find the line for http_requests_total
	lines := strings.Split(bodyStr, "\n")
	var counterLine string
	for _, line := range lines {
		if strings.HasPrefix(line, "http_requests_total{") {
			counterLine = line
			break
		}
	}
	if counterLine == "" {
		t.Fatalf("no http_requests_total line found in metrics:\n%s", bodyStr)
	}

	// Required label‑value pairs
	required := []string{
		`service="test-service"`,
		`method="POST"`,
		`route="posts"`,
		`status="201"`,
		`status_group="2xx"`,
		` 1`, // value
	}
	for _, need := range required {
		if !strings.Contains(counterLine, need) {
			t.Fatalf("counter line missing %q:\n%s", need, counterLine)
		}
	}

	// Ensure raw path ID is not present anywhere
	if strings.Contains(bodyStr, "post-123") {
		t.Fatalf("metrics body contains raw path id:\n%s", bodyStr)
	}

	// Check active gauge
	if !strings.Contains(bodyStr, `http_requests_active{service="test-service"}`) {
		t.Fatalf("missing active request gauge:\n%s", bodyStr)
	}

	// Check service_operations_total exists
	if !strings.Contains(bodyStr, "service_operations_total") {
		t.Fatalf("missing service operation counter:\n%s", bodyStr)
	}

	// Verify histogram buckets
	requiredBuckets := []string{"0.1", "0.3", "0.5", "1", "2", "5"}
	for _, le := range requiredBuckets {
		if !strings.Contains(bodyStr, fmt.Sprintf(`le="%s"`, le)) {
			t.Errorf("missing histogram bucket le=%s in metrics body:\n%s", le, bodyStr)
		}
	}

	// Verify status_group label appears somewhere
	if !strings.Contains(bodyStr, `status_group="2xx"`) {
		t.Errorf("missing status_group=\"2xx\" label in metrics body:\n%s", bodyStr)
	}
}

func TestMetricsErrorGrouping(t *testing.T) {
	metricsServer := httptest.NewServer(promhttp.Handler())
	defer metricsServer.Close()

	handler := Metrics("test-service")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/posts/123", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	resp, err := http.Get(metricsServer.URL)
	if err != nil {
		t.Fatalf("failed to fetch metrics: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read metrics body: %v", err)
	}
	bodyStr := string(bodyBytes)

	// status 500
	lines := strings.Split(bodyStr, "\n")
	var counterLine string
	for _, line := range lines {
		if strings.HasPrefix(line, "http_requests_total{") && strings.Contains(line, `status="500"`) {
			counterLine = line
			break
		}
	}
	if counterLine == "" {
		t.Fatalf("no http_requests_total line with status 500 found in metrics:\n%s", bodyStr)
	}

	required := []string{
		`service="test-service"`,
		`method="GET"`,
		`route="posts"`,
		`status="500"`,
		`status_group="5xx"`,
		` 1`,
	}
	for _, need := range required {
		if !strings.Contains(counterLine, need) {
			t.Fatalf("counter line missing %q:\n%s", need, counterLine)
		}
	}

	// Also check raw status code appears somewhere else (already covered above)
	if !strings.Contains(bodyStr, `status="500"`) {
		t.Errorf("expected status=\"500\" in metrics body:\n%s", bodyStr)
	}
}
