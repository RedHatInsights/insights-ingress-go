package s3

import "github.com/aws/aws-sdk-go/aws/session"

// Stager provides the mechanism to stage a payload via aws S3
type Stager struct {
	Bucket   string
	Rejected string
	Sess     *session.Session
}
