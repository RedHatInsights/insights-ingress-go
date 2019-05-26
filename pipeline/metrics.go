package pipeline

import (
	"time"

	. "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	stageElapsed = promauto.NewHistogramVec(HistogramOpts{
		Name: "ingress_stage_seconds",
		Help: "Number of seconds spent waiting on stage",
	}, []string{})
	validationElapsed = promauto.NewHistogramVec(HistogramOpts{
		Name: "ingress_validate_submit_seconds",
		Help: "Number of seconds spent submitting to validation",
	}, []string{"validation"})
)

func observeStageElapsed(elapsed time.Duration) {
	stageElapsed.With(Labels{}).Observe(elapsed.Seconds())
}

func observeValidationElapsed(timestamp time.Time, validation string) {
	validationElapsed.With(Labels{
		"validation": validation,
	}).Observe(time.Since(timestamp).Seconds())
}
