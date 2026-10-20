package kafka

import (
	"time"

	p "github.com/prometheus/client_golang/prometheus"
	pa "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	payloadsProcessed = pa.NewCounterVec(p.CounterOpts{
		Name: "ingress_processed_payloads",
		Help: "The total number of processed events",
	}, []string{"outcome"})

	validationElapsed = pa.NewHistogramVec(p.HistogramOpts{
		Name: "ingress_validate_elapsed_seconds",
		Help: "Number of seconds spent to validating",
	}, []string{"outcome"})

	messageProduced = pa.NewCounterVec(p.CounterOpts{
		Name: "ingress_message_produced",
		Help: "The total number of messages produced",
	}, []string{"service"})
)

func inc(outcome string) {
	payloadsProcessed.With(p.Labels{"outcome": outcome}).Inc()
}

func incMessageProduced(service string) {
	messageProduced.With(p.Labels{"service": service}).Inc()
}

func observeValidationElapsed(timestamp time.Time, outcome string) {
	validationElapsed.With(p.Labels{
		"outcome": outcome,
	}).Observe(time.Since(timestamp).Seconds())
}
