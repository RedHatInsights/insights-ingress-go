package stage

import "io"

// Input contains data and metadata to be staged
type Input struct {
	Reader   io.Reader
	Key      string
	Metadata io.Reader
}

// Stager provides the mechanism to stage a payload
type Stager interface {
	Stage(*Input) (string, error)
	Reject(rawurl string) error
}