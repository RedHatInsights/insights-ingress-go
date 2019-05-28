package kafka

import (
	"time"

	p "github.com/prometheus/client_golang/prometheus"
	pa "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	validationElapsed = pa.NewHistogramVec(p.HistogramOpts{
		Name: "ingress_validate_elapsed_seconds",
		Help: "Number of seconds spent to validating",
	}, []string{"outcome"})
)

func observeValidationElapsed(timestamp time.Time, outcome string) {
	validationElapsed.With(p.Labels{
		"outcome": outcome,
	}).Observe(time.Since(timestamp).Seconds())
}
