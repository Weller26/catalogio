package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests processed by catalogio",
		},
		[]string{"method", "path", "status"},
	)

	HTTPRequestsDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "http_requests_duration_seconds",
			Help: "Histogram of response latency for HTTP requests",
			Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5},
		},
		[]string{"method", "path", "status"},
	)

	AuthAttemptsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_attempts_total",
			Help: "Total count of auth attempts",
		},
		[]string{"status"},
	)

	UsersRegisteredTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "users_registered_total",
			Help: "Total count of registered users",
		},
	)

	ItemsCreatedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "items_created_total",
			Help: "Total count of items created in the catalog",
		},
	)

	EntitiesCurrent = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "entities_current",
			Help: "Current number of entities in the database",
		},
		[]string{"type"},
	)
)