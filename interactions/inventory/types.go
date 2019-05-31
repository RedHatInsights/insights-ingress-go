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
	IPAddresses  []string `json:"ip_addresses"`
	Account      string   `json:"account"`
	InsightsID   string   `json:"insights_id"`
	MachineID    string   `json:"machine_id"`
	SubManID     string   `json:"subscription_manager_id"`
	MacAddresses []string `json:"mac_addresses"`
	FQDN         string   `json:"fqdn"`
}
