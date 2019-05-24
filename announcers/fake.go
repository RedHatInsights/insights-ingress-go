package announcers

import (
	l "github.com/redhatinsights/insights-ingress-go/logger"
	"github.com/redhatinsights/insights-ingress-go/validators"
	"go.uber.org/zap"
)

// Fake is a fake announcer
type Fake struct {
	Event *validators.Response
}

// Announce does nothing
func (f *Fake) Announce(e *validators.Response) {
	f.Event = e
	l.Log.Info("Announce called", zap.String("request_id", e.RequestID))
}