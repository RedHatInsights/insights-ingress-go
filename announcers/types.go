package announcers

import "github.com/redhatinsights/insights-ingress-go/validators"

// Announcer for messages
type Announcer interface {
	Announce(e *validators.Response)
	Status(e *validators.Status)
	Stop()
}

// StatusAnnouncer is for payload-status announcements
type StatusAnnouncer interface {
	Announce(e *validators.Status)
	Stop()
}
