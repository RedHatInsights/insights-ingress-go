package telemetry

import (
	"context"

	"github.com/redhatinsights/platform-go-middlewares/v2/identity"
	"github.com/redhatinsights/platform-go-middlewares/v2/request_id"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type IngressAttributeProcessor struct{}

func (p *IngressAttributeProcessor) OnStart(ctx context.Context, s sdktrace.ReadWriteSpan) {
	attrs := []attribute.KeyValue{attribute.String("rh.service", "ingress")}
	if reqID := request_id.GetReqID(ctx); reqID != "" {
		attrs = append(attrs, attribute.String("rh.request_id", reqID))
	}
	if id := identity.GetIdentity(ctx); id.Identity.OrgID != "" {
		attrs = append(attrs, attribute.String("rh.org_id", id.Identity.OrgID))
	}
	s.SetAttributes(attrs...)
}

func (p *IngressAttributeProcessor) OnEnd(_ sdktrace.ReadOnlySpan)     {}
func (p *IngressAttributeProcessor) Shutdown(_ context.Context) error   { return nil }
func (p *IngressAttributeProcessor) ForceFlush(_ context.Context) error { return nil }
