package service

import (
	"testing"
	"time"

	dto "github.com/prometheus/client_model/go"
)

func TestObserveBusinessOperationRecordsNotificationMetric(t *testing.T) {
	counter := businessOperationTotal.WithLabelValues("notification-service", "get_notifications", businessStatusFailure)
	before := counterValue(t, counter)

	observeBusinessOperation("get_notifications", time.Now().Add(-time.Millisecond), businessStatusFailure)

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
