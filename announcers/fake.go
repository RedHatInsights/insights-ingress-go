package announcers

import (
	l "github.com/redhatinsights/insights-ingress-go/logger"
	"github.com/redhatinsights/insights-ingress-go/validators"
	"go.uber.org/zap"
)

// Fake is a fake announcer
type Fake struct {
	Event          *validators.Response
	StatusEvent    *Status
	AnnounceCalled bool
	StatusCalled   bool
	StopCalled     bool
}

// Announce does nothing
func (f *Fake) Announce(e *validators.Response) {
	f.AnnounceCalled = true
	f.Event = e
	l.Log.Info("Announce called", zap.String("request_id", e.RequestID))
}

// Status does nothing
func (f *Fake) Status(e *Status) {
	f.StatusCalled = true
	f.StatusEvent = e
	l.Log.Info("Announce called for Status", zap.String("request_id", e.RequestID))
}

// Stop does nothing
func (f *Fake) Stop() {
	f.StopCalled = true
	l.Log.Info("Stop called")
}
