package telemetry

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Serve basic telemetry info on given address.
// Right now it serves health checks and prometheus metrics.
func Serve(addr string) error {
	s := http.NewServeMux()
	s.HandleFunc("/healthz", func(rw http.ResponseWriter, _ *http.Request) {
		rw.WriteHeader(http.StatusOK)
	})
	s.HandleFunc("/readiness", func(rw http.ResponseWriter, _ *http.Request) {
		rw.WriteHeader(http.StatusOK)
	})
	s.Handle("/metrics", promhttp.Handler())

	// TODO: in future pprof info can be added here.
	return http.ListenAndServe(addr, s)
}
