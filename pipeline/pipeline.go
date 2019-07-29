package pipeline

import (
	"context"
	"time"

	"github.com/redhatinsights/insights-ingress-go/announcers"
	l "github.com/redhatinsights/insights-ingress-go/logger"
	"go.uber.org/zap"
)

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
		ps := &announcers.Status{
			Account:     ev.Account,
			RequestID:   ev.RequestID,
			Status:      "validated",
			StatusMsg:   "Payload validated by service",
			InventoryID: ev.ID,
		}
		l.Log.Info("Validation status received for payload", zap.String("request_id", ev.RequestID))
		p.Tracker.Status(ps)
		p.Announcer.Announce(ev)
		ps.Status = "announced"
		ps.StatusMsg = "Announced to platform"
		ps.Date = time.Now().UTC()
		p.Tracker.Status(ps)
	case iev, ok := <-p.InvalidChan:
		if !ok {
			return false
		}
		ps := &announcers.Status{
			Account:   iev.Account,
			RequestID: iev.RequestID,
			Status:    "Rejected",
			StatusMsg: "Payload not valid. rejecting",
		}
		l.Log.Info("Rejecting invalid payload", zap.String("request_id", iev.RequestID))
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
