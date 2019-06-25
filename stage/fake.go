package stage

import (
	"errors"
	"fmt"
	"time"
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

// Simulation is for fake input in testing
type Simulation struct {
	Input *Input
	Delay time.Duration
}

// Stage allows for simulated input
func (s *Simulation) Stage(input *Input) (string, error) {
	time.Sleep(s.Delay)
	return fmt.Sprintf("https://example.com/%s", input.Key), nil
}

// Reject is used for simulated rejections
func (s *Simulation) Reject(requestID string) error {
	time.Sleep(s.Delay)
	return nil
}

// GetURL is used for simulated url generation
func (s *Simulation) GetURL(requestID string) (string, error) {
	time.Sleep(s.Delay)
	return fmt.Sprintf("https://example.com/%s", requestID), nil
}
