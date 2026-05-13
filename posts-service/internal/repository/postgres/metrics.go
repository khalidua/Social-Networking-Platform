package postgres

import (
	"database/sql"
	"errors"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	dbQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "PostgreSQL query duration by service and repository operation.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "operation"},
	)
	dbErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_errors_total",
			Help: "PostgreSQL dependency errors by service and repository operation.",
		},
		[]string{"service", "operation"},
	)
)

func observeDBOperation(operation string, started time.Time, err error) {
	dbQueryDuration.WithLabelValues("posts-service", operation).Observe(time.Since(started).Seconds())
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		dbErrorsTotal.WithLabelValues("posts-service", operation).Inc()
	}
}
