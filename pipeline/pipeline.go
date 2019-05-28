package pipeline

import (
	"context"
	"time"

	l "github.com/redhatinsights/insights-ingress-go/logger"
	"github.com/redhatinsights/insights-ingress-go/stage"
	"github.com/redhatinsights/insights-ingress-go/validators"
	"go.uber.org/zap"
)

// Submit accepts a stage request and a validation request
func (p *Pipeline) Submit(in *stage.Input, vr *validators.Request) {
	defer in.Close()
	start := time.Now()
	url, err := p.Stager.Stage(in)
	observeStageElapsed(time.Since(start))
	if err != nil {
		l.Log.Error("Error staging", zap.String("key", in.Key), zap.Error(err))
		return
	}
	vr.URL = url
	vr.Timestamp = time.Now()
	p.Validator.Validate(vr)
}

// Tick is one loop iteration that handles post-validation activities
func (p *Pipeline) Tick(ctx context.Context) bool {
	select {
	case ev := <-p.ValidChan:
		p.Announcer.Announce(ev)
	case iev := <-p.InvalidChan:
		p.Stager.Reject(iev.RequestID)
	case <-ctx.Done():
		return false
	}
	return true
}

// Start loops forever until Tick is canceled
func (p *Pipeline) Start(ctx context.Context) {
	for p.Tick(ctx) {
	}
}
