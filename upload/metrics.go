package upload

import (
	. "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	requests = promauto.NewCounterVec(CounterOpts{
		Name: "ingress_requests",
		Help: "Total number of POSTs to /upload",
	}, []string{"useragent"})

	payloadSize = promauto.NewHistogramVec(HistogramOpts{
		Name: "ingress_payload_sizes",
		Help: "Size of payloads posted",
		Buckets: []float64{
			1024 * 10,
			1024 * 100,
			1024 * 1000,
			1024 * 10000,
		},
	}, []string{"useragent"})
)

func incRequests(userAgent string) {
	requests.With(Labels{"useragent": userAgent}).Inc()
}

func observeSize(userAgent string, size int64) {
	payloadSize.With(Labels{"useragent": userAgent}).Observe(float64(size))
}
