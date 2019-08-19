package upload

import (
	"strings"
	"time"

	p "github.com/prometheus/client_golang/prometheus"
	pa "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	requests = pa.NewCounterVec(p.CounterOpts{
		Name: "ingress_requests",
		Help: "Total number of POSTs to /upload",
	}, []string{"useragent"})

	payloadSize = pa.NewHistogramVec(p.HistogramOpts{
		Name: "ingress_payload_sizes",
		Help: "Size of payloads posted",
		Buckets: []float64{
			1024 * 10,
			1024 * 100,
			1024 * 1000,
			1024 * 10000,
		},
	}, []string{"useragent"})

	stageElapsed = pa.NewHistogramVec(p.HistogramOpts{
		Name: "ingress_stage_seconds",
		Help: "Number of seconds spent waiting on stage",
	}, []string{})
)

func incRequests(userAgent string) {
	requests.With(p.Labels{"useragent": NormalizeUserAgent(userAgent)}).Inc()
}

func observeSize(userAgent string, size int64) {
	payloadSize.With(p.Labels{"useragent": NormalizeUserAgent(userAgent)}).Observe(float64(size))
}

func observeStageElapsed(elapsed time.Duration) {
	stageElapsed.With(p.Labels{}).Observe(elapsed.Seconds())
}

// NormalizeUserAgent removes high-cardinality information from user agent strings
func NormalizeUserAgent(userAgent string) string {
	if strings.Contains(userAgent, "-operator") {
		return strings.Fields(userAgent)[0]
	}
	return userAgent
}
