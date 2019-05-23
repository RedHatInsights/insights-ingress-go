package stage

import "errors"

type Fake struct {
	Input        *Input
	StageCalled  bool
	RejectCalled bool
	RejectedURL  string
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

func (f *Fake) Reject(rawurl string) error {
	f.RejectCalled = true
	f.RejectedURL = rawurl
	return nil
}
