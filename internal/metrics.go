package internal

import (
	"math"
	"net/url"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type ClientFetchMetrics struct {
	ResponseLatency    *prometheus.HistogramVec
	ResponseStatusCode *prometheus.CounterVec
}

func NewClientFetchMetrics(reg prometheus.Registerer) *ClientFetchMetrics {
	m := &ClientFetchMetrics{
		ResponseLatency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "fuel_prices_govuk_api_http_response_latency_seconds",
				Help:    "GOV.UK fuel finder client API HTTP response latency in seconds.",
				Buckets: []float64{0.1, 0.25, 0.5, 1, 2, 5, 10, 30, math.Inf(1)},
			},
			[]string{"path", "method"},
		),
		ResponseStatusCode: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "fuel_prices_govuk_api_http_response_status_codes_total",
				Help: "GOV.UK fuel finder client API total number of HTTP responses by status code.",
			},
			[]string{"path", "method", "status_code"},
		),
	}

	if reg != nil {
		reg.MustRegister(m.ResponseLatency)
		reg.MustRegister(m.ResponseStatusCode)
	}

	return m
}

func (m *ClientFetchMetrics) Record(start time.Time, method, endpoint string, statusCode int, err error) {
	if m != nil {
		u, _ := url.Parse(endpoint)
		m.ResponseLatency.WithLabelValues(u.Path, method).Observe(time.Since(start).Seconds())
		if err == nil {
			m.ResponseStatusCode.WithLabelValues(u.Path, method, strconv.Itoa(statusCode)).Inc()
		}
	}
}
