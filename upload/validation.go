package upload

import (
	"regexp"
	"errors"
	"github.com/redhatinsights/insights-ingress-go/validators"
)

var contentTypePat = regexp.MustCompile(`application/vnd\.redhat\.([a-z0-9-]+)\.([a-z0-9-]+).*`)

func getServiceDescriptor(contentType string) (*validators.ServiceDescriptor, error) {
	if contentType == "application/x-gzip; charset=binary" {
		return &validators.ServiceDescriptor{
			Service: "advisor",
			Category: "upload",
		}, nil
	}
	m := contentTypePat.FindStringSubmatch(contentType)
	if m == nil {
		return nil, errors.New("Failed to match on Content-Type: " + contentType)
	}
	return &validators.ServiceDescriptor{
		Service:  m[1],
		Category: m[2],
	}, nil
}