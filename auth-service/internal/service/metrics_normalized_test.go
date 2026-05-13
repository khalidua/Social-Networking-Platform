package service

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestObserveBusinessOperationRecordsNormalizedMetric(t *testing.T) {
	before := testutil.ToFloat64(normalizedBusinessOperationTotal.WithLabelValues("auth-service", operationAuthenticateUser, statusSuccess))

	observeBusinessOperation(operationAuthenticateUser, time.Now().Add(-time.Millisecond), statusSuccess)

	after := testutil.ToFloat64(normalizedBusinessOperationTotal.WithLabelValues("auth-service", operationAuthenticateUser, statusSuccess))
	if after != before+1 {
		t.Fatalf("normalized business operation counter delta = %v, want 1", after-before)
	}
	if count := testutil.CollectAndCount(normalizedBusinessOperationDuration, "business_operation_duration_seconds"); count == 0 {
		t.Fatal("expected normalized business duration metric to be collected")
	}
}
