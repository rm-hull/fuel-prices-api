package metrics

import (
	"math"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type SqlMetrics struct {
	QueryLatency *prometheus.HistogramVec
}

func NewSqlMetrics(reg prometheus.Registerer) *SqlMetrics {
	m := &SqlMetrics{
		QueryLatency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "fuel_prices_sql_query_latency_seconds",
				Help:    "GOV.UK fuel finder SQL query latency in seconds.",
				Buckets: []float64{0.1, 0.25, 0.5, 1, 2, 5, 10, 30, math.Inf(1)},
			},
			[]string{"query"},
		),
	}

	RegisterOrPanic(reg, m.QueryLatency)

	return m
}

func (m *SqlMetrics) Record(start time.Time, query string) {
	m.QueryLatency.WithLabelValues(query).Observe(time.Since(start).Seconds())
}
