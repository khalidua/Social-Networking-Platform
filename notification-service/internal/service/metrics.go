package service

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	businessStatusSuccess = "success"
	businessStatusFailure = "failure"
)

var (
	businessOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "business_operation_duration_seconds",
			Help:    "Execution time of service-level business operations.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "operation"},
	)
	businessOperationTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "business_operation_total",
			Help: "Total service-level business operations partitioned by operation and outcome.",
		},
		[]string{"service", "operation", "status"},
	)
)

func observeBusinessOperation(operation string, started time.Time, status string) {
	businessOperationDuration.WithLabelValues("notification-service", operation).Observe(time.Since(started).Seconds())
	businessOperationTotal.WithLabelValues("notification-service", operation, status).Inc()
}
