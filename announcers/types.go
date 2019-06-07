package announcers

import "github.com/redhatinsights/insights-ingress-go/validators"

// Announcer
type Announcer interface {
	Announce(e *validators.Response)
	Stop()
}
