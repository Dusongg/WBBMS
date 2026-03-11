package observability

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bookadmin_http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"method", "route", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "bookadmin_http_request_duration_seconds",
			Help:    "Duration of HTTP requests.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "route"},
	)

	workerProcessedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bookadmin_worker_processed_total",
			Help: "Total number of processed worker messages.",
		},
		[]string{"stream", "result"},
	)

	workerFailuresTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bookadmin_worker_failures_total",
			Help: "Total number of worker processing failures.",
		},
		[]string{"stream", "stage"},
	)

	workerPendingGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "bookadmin_worker_pending_messages",
			Help: "Current pending messages in Redis Stream consumer groups.",
		},
		[]string{"stream"},
	)

	dbOpenConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "bookadmin_db_open_connections",
			Help: "Current number of open database connections.",
		},
	)

	dbIdleConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "bookadmin_db_idle_connections",
			Help: "Current number of idle database connections.",
		},
	)

	dbInUseConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "bookadmin_db_in_use_connections",
			Help: "Current number of in-use database connections.",
		},
	)

	redisPingDuration = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "bookadmin_redis_ping_duration_ms",
			Help: "Latest Redis ping latency in milliseconds.",
		},
	)
)

func RecordHTTPRequest(method, route string, status int, duration time.Duration) {
	httpRequestsTotal.WithLabelValues(method, route, strconv.Itoa(status)).Inc()
	httpRequestDuration.WithLabelValues(method, route).Observe(duration.Seconds())
}

func AddWorkerProcessed(stream, result string, count int) {
	workerProcessedTotal.WithLabelValues(stream, result).Add(float64(count))
}

func IncWorkerFailure(stream, stage string) {
	workerFailuresTotal.WithLabelValues(stream, stage).Inc()
}

func SetWorkerPending(stream string, pending int64) {
	workerPendingGauge.WithLabelValues(stream).Set(float64(pending))
}

func SetDBStats(open, idle, inUse int) {
	dbOpenConnections.Set(float64(open))
	dbIdleConnections.Set(float64(idle))
	dbInUseConnections.Set(float64(inUse))
}

func SetRedisPing(duration time.Duration) {
	redisPingDuration.Set(float64(duration.Milliseconds()))
}
