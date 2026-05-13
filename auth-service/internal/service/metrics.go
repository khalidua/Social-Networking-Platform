package service

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	operationAuthenticateUser = "authenticate_user"
	statusSuccess             = "success"
	statusFailure             = "failure"
)

var (
	businessOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "auth_service",
			Name:      "business_operation_duration_seconds",
			Help:      "Execution time of auth-service business operations.",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"operation"},
	)
	businessOperationTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "auth_service",
			Name:      "business_operation_total",
			Help:      "Total number of auth-service business operations partitioned by outcome.",
		},
		[]string{"status"},
	)
)

func observeBusinessOperation(operation string, started time.Time, status string) {
	businessOperationDuration.WithLabelValues(operation).Observe(time.Since(started).Seconds())
	businessOperationTotal.WithLabelValues(status).Inc()
}
