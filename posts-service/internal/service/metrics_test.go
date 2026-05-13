package service

import (
	"testing"
	"time"

	dto "github.com/prometheus/client_model/go"
)

func TestObserveBusinessOperationRecordsFailureAndDuration(t *testing.T) {
	counter := businessOperationTotal.WithLabelValues("posts-service", "create_post", businessStatusFailure)
	before := counterValue(t, counter)

	observeBusinessOperation("create_post", time.Now().Add(-time.Millisecond), businessStatusFailure)

	after := counterValue(t, counter)
	if after != before+1 {
		t.Fatalf("business operation counter delta = %v, want 1", after-before)
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
