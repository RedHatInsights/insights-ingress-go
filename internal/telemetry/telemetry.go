package telemetry

import (
	"context"
	"os"
	"time"

	"github.com/redhatinsights/insights-ingress-go/internal/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.41.0"
)

func InitTracer(cfg config.OtelConfig) (func(context.Context) error, error) {
	noop := func(context.Context) error { return nil }

	if !cfg.Enabled {
		return noop, nil
	}

	ctx := context.Background()

	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(cfg.Endpoint),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return noop, err
	}

	version := os.Getenv("IMAGE_TAG")
	if version == "" {
		version = "unknown"
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(version),
		),
	)
	if err != nil {
		return noop, err
	}

	sampler := sdktrace.ParentBased(
		sdktrace.TraceIDRatioBased(cfg.SamplingRate),
	)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter,
			sdktrace.WithMaxQueueSize(cfg.BSPMaxQueueSize),
			sdktrace.WithMaxExportBatchSize(cfg.BSPMaxExportBatchSize),
			sdktrace.WithBatchTimeout(time.Duration(cfg.BSPScheduleDelay)*time.Millisecond),
			sdktrace.WithExportTimeout(time.Duration(cfg.BSPExportTimeout)*time.Millisecond),
		),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
		sdktrace.WithSpanProcessor(&IngressAttributeProcessor{}),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return tp.Shutdown, nil
}
