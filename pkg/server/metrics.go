package server

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	metricQueries = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "doqd_queries",
		Help: "Total queries",
	})
	metricValidQueries = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "doqd_valid_queries",
		Help: "Total valid queries",
	})
	metricUpstreamErrors = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "doqd_upstream_errors",
		Help: "Total upstream errors",
	})
)

// MetricsListen starts the metrics HTTP server
func MetricsListen(listenAddr string) error {
	http.Handle("/metrics", promhttp.Handler())
	return http.ListenAndServe(listenAddr, nil)
}
