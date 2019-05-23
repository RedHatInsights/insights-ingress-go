package pipeline

import (
	"context"

	l "github.com/redhatinsights/insights-ingress-go/logger"
	"github.com/redhatinsights/insights-ingress-go/stage"
	"github.com/redhatinsights/insights-ingress-go/validators"
	"go.uber.org/zap"
)

// Submit accepts a stage request and a validation request
func (p *Pipeline) Submit(in *stage.Input, vr *validators.Request) {
	url, err := p.Stager.Stage(in)
	if err != nil {
		l.Log.Error("Error staging", zap.String("key", in.Key), zap.Error(err))
		return
	}
	vr.URL = url
	p.Validator.Validate(vr)
}

// Tick handles one loop for handling post-validation activities
func (p *Pipeline) Tick(ctx context.Context) bool {
	select {
	case ev := <-p.ValidChan:
		inc("success")
		p.Announcer.Announce(ev)
	case iev := <-p.InvalidChan:
		inc("failure")
		p.Stager.Reject(iev.RequestID)
	case <-ctx.Done():
		return false
	}
	return true
}

// Start watches the announcer channel for new events and calls announce
func (p *Pipeline) Start(ctx context.Context) {
	for {
		if !p.Tick(ctx) {
			return
		}
	}
}
