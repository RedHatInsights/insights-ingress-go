package telemetry_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/sirupsen/logrus"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	"github.com/redhatinsights/insights-ingress-go/internal/telemetry"
)

var _ = Describe("TraceLogHook", func() {
	var hook *telemetry.TraceLogHook

	BeforeEach(func() {
		hook = &telemetry.TraceLogHook{}
	})

	It("should return all log levels", func() {
		Expect(hook.Levels()).To(Equal(logrus.AllLevels))
	})

	Context("when entry has no context", func() {
		It("should not add trace fields", func() {
			entry := &logrus.Entry{
				Logger: logrus.New(),
				Data:   logrus.Fields{},
			}
			err := hook.Fire(entry)
			Expect(err).To(BeNil())
			Expect(entry.Data).ToNot(HaveKey("trace_id"))
			Expect(entry.Data).ToNot(HaveKey("span_id"))
		})
	})

	Context("when entry has context without a span", func() {
		It("should not add trace fields", func() {
			entry := &logrus.Entry{
				Logger:  logrus.New(),
				Data:    logrus.Fields{},
				Context: context.Background(),
			}
			err := hook.Fire(entry)
			Expect(err).To(BeNil())
			Expect(entry.Data).ToNot(HaveKey("trace_id"))
			Expect(entry.Data).ToNot(HaveKey("span_id"))
		})
	})

	Context("when entry has context with an active span", func() {
		It("should inject trace_id and span_id", func() {
			tp := sdktrace.NewTracerProvider()
			defer tp.Shutdown(context.Background())
			tracer := tp.Tracer("test")
			ctx, span := tracer.Start(context.Background(), "test-span")
			defer span.End()

			sc := trace.SpanFromContext(ctx).SpanContext()

			entry := &logrus.Entry{
				Logger:  logrus.New(),
				Data:    logrus.Fields{},
				Context: ctx,
			}
			err := hook.Fire(entry)
			Expect(err).To(BeNil())
			Expect(entry.Data).To(HaveKey("trace_id"))
			Expect(entry.Data).To(HaveKey("span_id"))
			Expect(entry.Data["trace_id"]).To(Equal(sc.TraceID().String()))
			Expect(entry.Data["span_id"]).To(Equal(sc.SpanID().String()))
		})
	})
})
