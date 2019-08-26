package stage

import "io"

// Input contains data and metadata to be staged
type Input struct {
	Payload io.ReadCloser
	Key     string
}

// Close closes the underlying ReadCloser as long as it isn't nil
func (i *Input) Close() {
	if i.Payload != nil {
		i.Payload.Close()
	}
}

// Stager provides the mechanism to stage a payload
type Stager interface {
	Stage(*Input) (string, error)
	GetURL(requestID string) (string, error)
}
