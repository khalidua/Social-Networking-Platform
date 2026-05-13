package postgres

import (
	"database/sql"
	"testing"
	"time"

	dto "github.com/prometheus/client_model/go"
)

func TestObserveDBOperationSkipsNoRowsAsDependencyError(t *testing.T) {
	counter := dbErrorsTotal.WithLabelValues("posts-service", "select_post")
	before := counterValue(t, counter)

	observeDBOperation("select_post", time.Now().Add(-time.Millisecond), sql.ErrNoRows)

	after := counterValue(t, counter)
	if after != before {
		t.Fatalf("db error counter changed for sql.ErrNoRows: before=%v after=%v", before, after)
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
