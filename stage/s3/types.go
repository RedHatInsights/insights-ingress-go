package s3

import "github.com/aws/aws-sdk-go/aws/session"

// S3Stager provides the mechanism to stage a payload via aws S3
type S3Stager struct {
	Bucket   string
	Rejected string
	Sess     *session.Session
}
