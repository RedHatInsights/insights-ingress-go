package telemetry

import (
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
)

type TraceLogHook struct{}

func (h *TraceLogHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *TraceLogHook) Fire(entry *logrus.Entry) error {
	if entry.Context == nil {
		return nil
	}
	sc := trace.SpanFromContext(entry.Context).SpanContext()
	if sc.HasTraceID() {
		entry.Data["trace_id"] = sc.TraceID().String()
		entry.Data["span_id"] = sc.SpanID().String()
	}
	return nil
}
