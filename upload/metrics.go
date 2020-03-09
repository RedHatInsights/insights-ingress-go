package upload

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
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

	responseCodes = pa.NewCounterVec(p.CounterOpts{
		Name: "ingress_responses",
		Help: "Count of response codes by code and user-agent",
	}, []string{"useragent", "code"})

	insightsClientCoreRE = regexp.MustCompile(`(insights-client/[0-9.]+) \((Core [0-9.]+)`)
	insightsClientBareRE = regexp.MustCompile(`insights-client/[0-9.]+`)
	accessInsightsRE     = regexp.MustCompile(`redhat-access-insights/[0-9.]+`)
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

	if strings.Contains(userAgent, "Core") {
		submatches := insightsClientCoreRE.FindStringSubmatch(userAgent)
		if submatches != nil {
			return fmt.Sprintf("%s %s", submatches[1], submatches[2])
		}
	}

	if strings.Contains(userAgent, "insights-client") {
		m := insightsClientBareRE.FindString(userAgent)
		if m != "" {
			return m
		}
	}

	if strings.Contains(userAgent, "redhat-access-insights") {
		m := accessInsightsRE.FindString(userAgent)
		if m != "" {
			return m
		}
	}

	return userAgent
}

type metricTrackingResponseWriter struct {
	Wrapped   http.ResponseWriter
	UserAgent string
}

func (m *metricTrackingResponseWriter) Header() http.Header {
	return m.Wrapped.Header()
}

func (m *metricTrackingResponseWriter) Write(b []byte) (int, error) {
	return m.Wrapped.Write(b)
}

func (m *metricTrackingResponseWriter) WriteHeader(statusCode int) {
	responseCodes.With(p.Labels{"useragent": NormalizeUserAgent(m.UserAgent), "code": strconv.Itoa(statusCode)}).Inc()
	m.Wrapped.WriteHeader(statusCode)
}

// ResponseMetricsMiddleware wraps the ResponseWriter such that metrics for each
// response type get tracked
func ResponseMetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ww := &metricTrackingResponseWriter{
			UserAgent: r.Header.Get("user-agent"),
			Wrapped:   w,
		}
		next.ServeHTTP(ww, r)
	})
}
