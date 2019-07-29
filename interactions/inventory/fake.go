package inventory

import (
	"github.com/redhatinsights/insights-ingress-go/validators"
)

// Fake structure to hold the ID
type Fake struct {
}

// GetID fake to get an ID
func (f *Fake) GetID(md validators.Metadata, account string, ident string) (string, error) {
	return "1234-abcd-5678-efgh", nil
}
