package stage

import (
	"errors"
)

// Fake is used for tests
type Fake struct {
	Input        *Input
	StageCalled  bool
	RejectCalled bool
	RequestID    string
	ShouldError  bool
	URL          string
}

// Stage is used for testing with Fake input
func (f *Fake) Stage(input *Input) (string, error) {
	f.Input = input
	f.StageCalled = true
	if f.ShouldError {
		return "", errors.New("Fake Stager Error")
	}
	return f.URL, nil
}

// Reject is used for testing fake rejections
func (f *Fake) Reject(requestID string) error {
	f.RejectCalled = true
	f.RequestID = requestID
	return nil
}

// GetURL is used to test fake url returns
func (f *Fake) GetURL(requestID string) (string, error) {
	return f.URL, nil
}
