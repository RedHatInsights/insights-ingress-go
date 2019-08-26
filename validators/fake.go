package validators

import (
	"errors"
)

// Fake allows for creation of testing objects
type Fake struct {
	In     *Request
	Called bool
}

func (v *Fake) Validate(in *Request) {
	v.Called = true
	v.In = in
}

// ValidateService allows for testing service validations
func (v *Fake) ValidateService(service *ServiceDescriptor) error {
	if service.Service == "failed" {
		return errors.New("failed is an invalid service")
	}
	return nil
}
