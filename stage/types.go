package stage

import (
	"io"

	"github.com/aws/aws-sdk-go/aws/session"
)

// Input contains data and metadata to be staged
type Input struct {
	Reader   io.Reader
	Key      string
	Metadata io.Reader
}

// Stager provides the mechanism to stage a payload
type Stager interface {
	Stage(*Input) (string, error)
}

// S3Stager provides the mechanism to stage a payload via aws S3
type S3Stager struct {
	Bucket string
	Sess   *session.Session
}
