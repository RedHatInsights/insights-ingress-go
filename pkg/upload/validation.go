package upload

import (
	"errors"
	"regexp"

	"github.com/redhatinsights/insights-ingress-go/pkg/validators"
)

var contentTypePat = regexp.MustCompile(`application/vnd\.redhat\.([a-z0-9-]+)\.([a-z0-9-]+).*`)

func getServiceDescriptor(contentType string) (*validators.ServiceDescriptor, error) {
	switch ctype := contentType; ctype {
	case "application/x-gzip; charset=binary", "application/gzip", "application/gzip; charset=binary":
		return &validators.ServiceDescriptor{
			Service:  "advisor",
			Category: "upload",
		}, nil
	default:
		m := contentTypePat.FindStringSubmatch(ctype)
		if m == nil {
			return nil, errors.New("Failed to match on Content-Type: " + ctype)
		}
		return &validators.ServiceDescriptor{
			Service:  m[1],
			Category: m[2],
		}, nil
	}
}
