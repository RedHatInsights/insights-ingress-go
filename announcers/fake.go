package announcers

import (
	"sync"

	l "github.com/redhatinsights/insights-ingress-go/logger"
	"go.uber.org/zap"
)

// Fake is a fake announcer
type Fake struct {
	StatusEvent     *Status
	AnnounceCalledV bool
	StatusCalledV   bool
	StopCalledV     bool
	lock            sync.Mutex
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
