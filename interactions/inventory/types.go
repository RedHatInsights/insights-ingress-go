package inventory

// Inventory is the JSON stucture of the inventory response
type Inventory struct {
	Data []struct {
		Status int `json:"status"`
		Host   struct {
			ID string `json:"id"`
		} `json:"host"`
	} `json:"data"`
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
}
