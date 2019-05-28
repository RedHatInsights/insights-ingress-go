package upload

import (
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
)

func incRequests(userAgent string) {
	requests.With(p.Labels{"useragent": userAgent}).Inc()
}

func observeSize(userAgent string, size int64) {
	payloadSize.With(p.Labels{"useragent": userAgent}).Observe(float64(size))
}
