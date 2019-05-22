package announcers

import "github.com/redhatinsights/insights-ingress-go/validators"

// Announcer
type Announcer interface {
	Announce(e *validators.Response)
}

// HostInfo represents details about managed hosts
type HostInfo struct {
	ID               string `json:"id"`
	SatelliteManaged string `json:"satellite_managed"`
}

// AvailableEvent is sent to the available topic for each validated payload
type AvailableEvent struct {
	Account     string            `json:"account"`
	B64Identity []byte            `json:"b64_identity"`
	RequestID   string            `json:"request_id"`
	Principal   string            `json:"principal"`
	Service     string            `json:"service"`
	URL         string            `json:"url"`
	Extras      map[string]string `json:"extras"`
}
