package announcers

import (
	l "github.com/redhatinsights/insights-ingress-go/logger"
	"github.com/redhatinsights/insights-ingress-go/validators"
	"go.uber.org/zap"
)

// Fake is a fake announcer
type Fake struct {
	Event       *validators.Response
	StatusEvent *validators.Status
}

// Announce does nothing
func (f *Fake) Announce(e *validators.Response) {
	f.Event = e
	l.Log.Info("Announce called", zap.String("request_id", e.RequestID))
}

// Status does nothing
func (f *Fake) Status(e *validators.Status) {
	f.StatusEvent = e
	l.Log.Info("Announce called for Status", zap.String("request_id", e.RequestID))
}

// Stop does nothing
func (f *Fake) Stop() {
	l.Log.Info("Stop called")
}
