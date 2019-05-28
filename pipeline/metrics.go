package pipeline

import (
	"time"

	p "github.com/prometheus/client_golang/prometheus"
	pa "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	stageElapsed = pa.NewHistogramVec(p.HistogramOpts{
		Name: "ingress_stage_seconds",
		Help: "Number of seconds spent waiting on stage",
	}, []string{})
)

func observeStageElapsed(elapsed time.Duration) {
	stageElapsed.With(p.Labels{}).Observe(elapsed.Seconds())
}
