package s3

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// Stager provides the mechanism to stage a payload via aws S3
type Stager struct {
	Bucket   string
	Sess     *session.Session
	Uploader *s3manager.Uploader
	Client   *s3.S3
}
