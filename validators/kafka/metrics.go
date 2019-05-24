package kafka

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)
var (
	payloadsProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ingress_processed_payloads",
		Help: "The total number of processed events",
	}, []string{"validation"})
)

func inc(value string) {
	payloadsProcessed.With(prometheus.Labels{"validation": value}).Inc()
}