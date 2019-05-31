package pipeline

import (
	"context"
	"time"

	i "github.com/redhatinsights/insights-ingress-go/interactions/inventory"
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
	if vr.Metadata != nil {
		vr.ID, err = i.PostInventory(vr)
		if err != nil {
			l.Log.Error("Unable to post to inventory", zap.Error(err),
				zap.String("request_id", vr.RequestID))
		}
	}
	l.Log.Info("Inventory ID: ", zap.String("ID", vr.ID))
	vr.Timestamp = time.Now()
	p.Validator.Validate(vr)
}

// Tick is one loop iteration that handles post-validation activities
func (p *Pipeline) Tick(ctx context.Context) bool {
	select {
	case ev, ok := <-p.ValidChan:
		if !ok {
			return false
		}
		p.Announcer.Announce(ev)
	case iev, ok := <-p.InvalidChan:
		if !ok {
			return false
		}
		p.Stager.Reject(iev.RequestID)
	case <-ctx.Done():
		return false
	}
	return true
}

// Start loops forever until Tick is canceled
func (p *Pipeline) Start(ctx context.Context, stopped chan struct{}) {
	for p.Tick(ctx) {
	}
	l.Log.Info("Tick returned false, closing stopped channel")
	close(stopped)
}
