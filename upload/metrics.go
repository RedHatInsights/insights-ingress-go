package upload

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	p "github.com/prometheus/client_golang/prometheus"
	pa "github.com/prometheus/client_golang/prometheus/promauto"
	l "github.com/redhatinsights/insights-ingress-go/logger"
	"go.uber.org/zap"
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

	responseCodes = make(map[int]*p.CounterVec)
)

func init() {
	codes := []int{
		http.StatusAccepted,
		http.StatusBadRequest,
		http.StatusInternalServerError,
		http.StatusRequestEntityTooLarge,
		http.StatusUnsupportedMediaType,
	}
	for code := range codes {
		responseCodes[code] = pa.NewCounterVec(p.CounterOpts{
			Name: fmt.Sprintf("ingress_response_%d", code),
			Help: fmt.Sprintf("Total Number of %d response codes by user-agent", code),
		}, []string{"useragent"})
	}
}

func incResponse(userAgent string, code int) {
	m, ok := responseCodes[code]
	if !ok {
		l.Log.Error("tried to inc a metric that does not exist.  Be sure to define it in upload/metrics.go.", zap.Int("code", code))
		return
	}
	m.With(p.Labels{"useragent": NormalizeUserAgent(userAgent)}).Inc()
}

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
