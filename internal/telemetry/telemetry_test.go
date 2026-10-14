package telemetry_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"

	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace/noop"

	"github.com/redhatinsights/insights-ingress-go/internal/config"
	"github.com/redhatinsights/insights-ingress-go/internal/telemetry"
)

var testLog = &logrus.Logger{Out: logrus.StandardLogger().Out, Level: logrus.FatalLevel, Formatter: &logrus.TextFormatter{}}

var _ = Describe("InitTracer", func() {
	Context("when OTel is disabled", func() {
		It("should return a no-op shutdown function and no middlewares", func() {
			cfg := config.OtelConfig{
				Enabled:      false,
				Endpoint:     "localhost:4318",
				SamplingRate: 1.0,
				ServiceName:  "ingress",
			}
			shutdown, middlewares, err := telemetry.InitTracer(cfg, testLog)
			Expect(err).To(BeNil())
			Expect(shutdown).ToNot(BeNil())
			Expect(middlewares).To(BeNil())

			err = shutdown(context.Background())
			Expect(err).To(BeNil())
		})

		It("should not set an SDK TracerProvider", func() {
			cfg := config.OtelConfig{Enabled: false}
			_, _, err := telemetry.InitTracer(cfg, testLog)
			Expect(err).To(BeNil())

			tp := otel.GetTracerProvider()
			_, isSDK := tp.(*sdktrace.TracerProvider)
			Expect(isSDK).To(BeFalse())
		})
	})

	Context("when OTel is enabled", func() {
		It("should set an SDK TracerProvider and return middlewares", func() {
			cfg := config.OtelConfig{
				Enabled:      true,
				Endpoint:     "localhost:4318",
				SamplingRate: 1.0,
				ServiceName:  "test-ingress",
			}
			shutdown, middlewares, err := telemetry.InitTracer(cfg, testLog)
			Expect(err).To(BeNil())
			Expect(shutdown).ToNot(BeNil())
			Expect(middlewares).To(HaveLen(2))

			tp := otel.GetTracerProvider()
			_, isSDK := tp.(*sdktrace.TracerProvider)
			Expect(isSDK).To(BeTrue())

			err = shutdown(context.Background())
			Expect(err).To(BeNil())

			otel.SetTracerProvider(noop.NewTracerProvider())
		})
	})
})
