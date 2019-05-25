package stage

import "io"

// Input contains data and metadata to be staged
type Input struct {
	Payload io.ReadCloser
	Key     string
}

// Stager provides the mechanism to stage a payload
type Stager interface {
	Stage(*Input) (string, error)
	Reject(requestID string) error
}
