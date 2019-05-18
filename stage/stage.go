package stage

import (
	"errors"
	"io"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
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

func getSession() *session.Session {
	return session.Must(session.NewSession())
}

// NewS3Stager constructs a new stager for the bucket
func NewS3Stager(bucket string) Stager {
	return &S3Stager{
		Bucket: bucket,
		Sess:   getSession(),
	}
}

// TODO: use context here? We want to store other things like user-agent and such...

// Stage stores the file in s3 and returns a presigned url
func (s *S3Stager) Stage(in *Input) (string, error) {
	uploader := s3manager.NewUploader(s.Sess)
	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(in.Key),
		Body:   in.Reader,
	})
	if err != nil {
		return "", errors.New("Failed to upload to s3: " + err.Error())
	}

	client := s3.New(s.Sess)
	req, _ := client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(in.Key),
	})
	url, err := req.Presign(24 * time.Hour)
	if err != nil {
		return "", errors.New("Failed to generate presigned url: " + err.Error())
	}

	return url, nil
}

func (s *S3Stager) copy(fromBucket string, fromKey string, toBucket string) error {
	src := fromBucket + "/" + fromKey
	client := s3.New(s.Sess)
	input := &s3.CopyObjectInput{
		Bucket:     aws.String(toBucket),
		CopySource: aws.String(src),
		Key:        aws.String(fromKey),
	}
	_, err := client.CopyObject(input)
	if err != nil {
		return errors.New("Failed to copy from " + src + " to " + toBucket)
	}
	return nil
}
