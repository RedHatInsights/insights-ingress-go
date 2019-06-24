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
	ps := &validators.Status{
		Account:   vr.Account,
		Service:   "ingress",
		RequestID: vr.RequestID,
		Status:    "processing",
		StatusMsg: "Sent to validation service",
		Date:      time.Now(),
	}
	p.Tracker.Status(ps)
	p.Validator.Validate(vr)
}

// Tick is one loop iteration that handles post-validation activities
func (p *Pipeline) Tick(ctx context.Context) bool {
	select {
	case ev, ok := <-p.ValidChan:
		if !ok {
			return false
		}
		url, err := p.Stager.GetURL(ev.RequestID)
		if err != nil {
			l.Log.Error("Failed to GetURL", zap.String("request_id", ev.RequestID), zap.Error(err))
			return true
		}
		ev.URL = url
		ps := &validators.Status{
			Account:     ev.Account,
			Service:     "ingress",
			RequestID:   ev.RequestID,
			Status:      "validated",
			StatusMsg:   "Payload validated by service",
			InventoryID: ev.ID,
			Date:        time.Now(),
		}
		p.Tracker.Status(ps)
		p.Announcer.Announce(ev)
		ps.Status = "announced"
		ps.StatusMsg = "Announced to platform"
		ps.Date = time.Now()
		p.Tracker.Status(ps)
	case iev, ok := <-p.InvalidChan:
		if !ok {
			return false
		}
		ps := &validators.Status{
			Account:   iev.Account,
			Service:   "ingress",
			RequestID: iev.RequestID,
			Status:    "Rejected",
			StatusMsg: "Payload not valid. rejecting",
			Date:      time.Now(),
		}
		p.Tracker.Status(ps)
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
