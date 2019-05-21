package announcers

import "cloud.redhat.com/ingress/validators"

// Fake is a fake announcer
type Fake struct {
	event *validators.Response
}

// Announce does nothing
func (f *Fake) Announce(e *validators.Response) {
	f.event = e
}
