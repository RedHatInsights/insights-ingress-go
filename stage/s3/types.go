package s3

import "github.com/aws/aws-sdk-go/aws/session"

// StageS3 provides the mechanism to stage a payload via aws S3
type StageS3 struct {
	Bucket   string
	Rejected string
	Sess     *session.Session
}
