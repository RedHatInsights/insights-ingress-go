package announcers

// Announcer 
type Announcer interface {
	Announce(e *AvailableEvent)	
}

// HostInfo represents details about managed hosts
type HostInfo struct {
	ID               string `json:"id"`
	SatelliteManaged string `json:"satellite_managed"`
}

// AvailableEvent is sent to the available topic for each validated payload
type AvailableEvent struct {
	Account     string   `json:"account"`
	B64Identity []byte   `json:"b64_identity"`
	Host        HostInfo `json:"host_info"`
	RequestID   string   `json:"request_id"`
	Principal   string   `json:"principal"`
	Service     string   `json:"service"`
	URL         string   `json:"url"`
}