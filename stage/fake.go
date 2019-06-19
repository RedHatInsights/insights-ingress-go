package stage

import (
	"errors"
	"fmt"
	"time"
)

type Fake struct {
	Input        *Input
	StageCalled  bool
	RejectCalled bool
	RequestID    string
	ShouldError  bool
	URL          string
}

func (f *Fake) Stage(input *Input) (string, error) {
	f.Input = input
	f.StageCalled = true
	if f.ShouldError {
		return "", errors.New("Fake Stager Error")
	}
	return f.URL, nil
}

func (f *Fake) Reject(requestID string) error {
	f.RejectCalled = true
	f.RequestID = requestID
	return nil
}

func (f *Fake) GetURL(requestID string) (string, error) {
	return f.URL, nil
}

type Simulation struct {
	Input *Input
	Delay time.Duration
}

func (s *Simulation) Stage(input *Input) (string, error) {
	time.Sleep(s.Delay)
	return fmt.Sprintf("https://example.com/%s", input.Key), nil
}

func (s *Simulation) Reject(requestID string) error {
	time.Sleep(s.Delay)
	return nil
}

func (s *Simulation) GetURL(requestID string) (string, error) {
	time.Sleep(s.Delay)
	return fmt.Sprintf("https://example.com/%s", requestID), nil
}
