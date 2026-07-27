package telemetry_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace/noop"

	"github.com/redhatinsights/insights-ingress-go/internal/config"
	"github.com/redhatinsights/insights-ingress-go/internal/telemetry"
)

var _ = Describe("InitTracer", func() {
	Context("when OTel is disabled", func() {
		It("should return a no-op shutdown function", func() {
			cfg := config.OtelConfig{
				Enabled:      false,
				Endpoint:     "localhost:4318",
				SamplingRate: 1.0,
				ServiceName:  "ingress",
			}
			shutdown, err := telemetry.InitTracer(cfg)
			Expect(err).To(BeNil())
			Expect(shutdown).ToNot(BeNil())

			err = shutdown(context.Background())
			Expect(err).To(BeNil())
		})

		It("should not set an SDK TracerProvider", func() {
			cfg := config.OtelConfig{Enabled: false}
			_, err := telemetry.InitTracer(cfg)
			Expect(err).To(BeNil())

			tp := otel.GetTracerProvider()
			_, isSDK := tp.(*sdktrace.TracerProvider)
			Expect(isSDK).To(BeFalse())
		})
	})

	Context("when OTel is enabled", func() {
		It("should set an SDK TracerProvider", func() {
			cfg := config.OtelConfig{
				Enabled:      true,
				Endpoint:     "localhost:4318",
				SamplingRate: 1.0,
				ServiceName:  "test-ingress",
			}
			shutdown, err := telemetry.InitTracer(cfg)
			Expect(err).To(BeNil())
			Expect(shutdown).ToNot(BeNil())

			tp := otel.GetTracerProvider()
			_, isSDK := tp.(*sdktrace.TracerProvider)
			Expect(isSDK).To(BeTrue())

			err = shutdown(context.Background())
			Expect(err).To(BeNil())

			otel.SetTracerProvider(noop.NewTracerProvider())
		})
	})
})
