package telemetry

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/redhatinsights/insights-ingress-go/internal/config"
	"github.com/redhatinsights/platform-go-middlewares/v2/identity"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.41.0"
)

func InitTracer(cfg config.OtelConfig, log *logrus.Logger) (func(context.Context) error, []func(http.Handler) http.Handler, error) {
	noop := func(context.Context) error { return nil }

	if !cfg.Enabled {
		return noop, nil, nil
	}

	ctx := context.Background()

	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(cfg.Endpoint),
	}
	if cfg.Insecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}

	exporter, err := otlptracehttp.New(ctx, opts...)
	if err != nil {
		return noop, nil, err
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
		return noop, nil, err
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

	middlewares := []func(http.Handler) http.Handler{
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if header := r.Header.Get("X-Rh-Identity"); header != "" {
					if id, err := identity.DecodeIdentity(header); err == nil {
						r = r.WithContext(identity.WithIdentity(r.Context(), id))
					}
				}
				next.ServeHTTP(w, r)
			})
		},
		otelhttp.NewMiddleware("ingress",
			otelhttp.WithFilter(func(r *http.Request) bool {
				return r.Method == "POST" && strings.HasSuffix(r.URL.Path, "/upload")
			}),
		),
	}

	log.WithFields(logrus.Fields{
		"endpoint":              cfg.Endpoint,
		"insecure":              cfg.Insecure,
		"sampling_rate":         cfg.SamplingRate,
		"service_name":          cfg.ServiceName,
		"bsp_max_queue_size":    cfg.BSPMaxQueueSize,
		"bsp_max_export_batch":  cfg.BSPMaxExportBatchSize,
		"bsp_schedule_delay_ms": cfg.BSPScheduleDelay,
		"bsp_export_timeout_ms": cfg.BSPExportTimeout,
	}).Info("OpenTelemetry tracing enabled")

	return tp.Shutdown, middlewares, nil
}
