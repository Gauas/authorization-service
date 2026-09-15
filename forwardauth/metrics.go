package forwardauth

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	requests     *prometheus.CounterVec
	denied       *prometheus.CounterVec
	latency      prometheus.Histogram
	redisLatency prometheus.Histogram
	jwksRefresh  prometheus.Counter
	jwksFailures prometheus.Counter
}

func NewMetrics(registerer prometheus.Registerer) *Metrics {
	metrics := &Metrics{
		requests:     prometheus.NewCounterVec(prometheus.CounterOpts{Name: "authorization_requests_total", Help: "ForwardAuth requests by status."}, []string{"status"}),
		denied:       prometheus.NewCounterVec(prometheus.CounterOpts{Name: "authorization_denied_total", Help: "Denied ForwardAuth requests by reason."}, []string{"reason"}),
		latency:      prometheus.NewHistogram(prometheus.HistogramOpts{Name: "authorization_latency_seconds", Help: "ForwardAuth decision latency.", Buckets: prometheus.DefBuckets}),
		redisLatency: prometheus.NewHistogram(prometheus.HistogramOpts{Name: "redis_revocation_lookup_seconds", Help: "Redis revocation lookup latency.", Buckets: prometheus.DefBuckets}),
		jwksRefresh:  prometheus.NewCounter(prometheus.CounterOpts{Name: "jwks_refresh_total", Help: "JWKS refresh attempts."}),
		jwksFailures: prometheus.NewCounter(prometheus.CounterOpts{Name: "jwks_refresh_failures_total", Help: "Failed JWKS refresh attempts."}),
	}
	registerer.MustRegister(metrics.requests, metrics.denied, metrics.latency, metrics.redisLatency, metrics.jwksRefresh, metrics.jwksFailures)
	return metrics
}

func (m *Metrics) ObserveRedisLookup(duration time.Duration) {
	m.redisLatency.Observe(duration.Seconds())
}

func (m *Metrics) ObserveJWKSRefresh(success bool) {
	m.jwksRefresh.Inc()
	if !success {
		m.jwksFailures.Inc()
	}
}

func (m *Metrics) Observe(status int, reason string, duration time.Duration) {
	m.requests.WithLabelValues(strconv.Itoa(status)).Inc()
	if status >= 400 {
		m.denied.WithLabelValues(reason).Inc()
	}
	m.latency.Observe(duration.Seconds())
}
