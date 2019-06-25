package inventory

import "github.com/redhatinsights/insights-ingress-go/validators"

// Response is the JSON stucture of the inventory response
type Response struct {
	Data []struct {
		Status int `json:"status"`
		Host   struct {
			ID string `json:"id"`
		} `json:"host"`
	} `json:"data"`
}

type facts struct {
	Namespace string                 `json:"namespace"`
	Facts     map[string]interface{} `json:"facts"`
}

// Metadata is the expected data from a client
type Metadata struct {
	IPAddresses  []string `json:"ip_addresses,omitempty"`
	Account      string   `json:"account,omitempty"`
	InsightsID   string   `json:"insights_id,omitempty"`
	MachineID    string   `json:"machine_id,omitempty"`
	SubManID     string   `json:"subscription_manager_id,omitempty"`
	MacAddresses []string `json:"mac_addresses,omitempty"`
	FQDN         string   `json:"fqdn,omitempty"`
	Facts        []facts  `json:"facts,omitempty"`
}

// Inventory can return an inventory ID
type Inventory interface {
	GetID(vr *validators.Request) (string, error)
}
