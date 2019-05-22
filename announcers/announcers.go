package announcers

import "github.com/redhatinsights/insights-ingress-go/validators"

// Fake is a fake announcer
type Fake struct {
	event *validators.Response
}

// Announce does nothing
func (f *Fake) Announce(e *validators.Response) {
	f.event = e
}
