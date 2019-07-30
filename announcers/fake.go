package announcers

import (
	"sync"

	l "github.com/redhatinsights/insights-ingress-go/logger"
	"github.com/redhatinsights/insights-ingress-go/validators"
	"go.uber.org/zap"
)

// Fake is a fake announcer
type Fake struct {
	Event           *validators.Response
	StatusEvent     *Status
	AnnounceCalledV bool
	StatusCalledV   bool
	StopCalledV     bool
	lock            sync.Mutex
}

// Announce does nothing
func (f *Fake) Announce(e *validators.Response) {
	f.lock.Lock()
	f.AnnounceCalledV = true
	f.lock.Unlock()
	f.Event = e
	l.Log.Info("Announce called", zap.String("request_id", e.RequestID))
}

// Status does nothing
func (f *Fake) Status(e *Status) {
	f.StatusCalledV = true
	f.StatusEvent = e
	l.Log.Info("Announce called for Status", zap.String("request_id", e.RequestID))
}

// Stop does nothing
func (f *Fake) Stop() {
	f.StopCalledV = true
	l.Log.Info("Stop called")
}

func (f *Fake) AnnounceCalled() bool {
	f.lock.Lock()
	defer f.lock.Unlock()
	return f.AnnounceCalledV
}
