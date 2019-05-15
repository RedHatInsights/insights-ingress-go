package upload

// ValidationRequest is sent to the validation topic for each new payload
type ValidationRequest struct {
	Account     string      `json:"account"`
	Principal   string      `json:"principal"`
	PayloadID   string      `json:"payload_id"`
	Size        int64       `json:"size"`
	Service     string      `json:"service"`
	Category    string      `json:"category"`
	B64Identity []byte      `json:"b64_identity"`
	Metadata    interface{} `json:"metadata"`
}

// HostInfo represents details about managed hosts
type HostInfo struct {
	ID               string `json:"id"`
	SatelliteManaged string `json:"satellite_managed"`
}

// AvailableEvent is sent to the available topic for each validated payload
type AvailableEvent struct {
	URL         string   `json:"url"`
	Service     string   `json:"service"`
	PayloadID   string   `json:"payload_id"`
	B64Identity []byte   `json:"b64_identity"`
	Account     string   `json:"account"`
	Principal   string   `json:"principal"`
	Host        HostInfo `json:"host_info"`
}
