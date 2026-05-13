package postgres

import (
	"errors"
	"testing"
	"time"

	dto "github.com/prometheus/client_model/go"
)

func TestObserveDBOperationRecordsLatencyAndErrors(t *testing.T) {
	counter := dbErrorsTotal.WithLabelValues("users-service", "select_user")
	before := counterValue(t, counter)

	observeDBOperation("select_user", time.Now().Add(-time.Millisecond), errors.New("database unavailable"))

	after := counterValue(t, counter)
	if after != before+1 {
		t.Fatalf("db error counter delta = %v, want 1", after-before)
	}
}

func counterValue(t *testing.T, metric interface{ Write(*dto.Metric) error }) float64 {
	t.Helper()
	var out dto.Metric
	if err := metric.Write(&out); err != nil {
		t.Fatalf("metric.Write: %v", err)
	}
	if out.Counter == nil {
		return 0
	}
	return out.Counter.GetValue()
}
