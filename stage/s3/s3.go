package s3

import (
	"errors"
	"time"

	"github.com/redhatinsights/insights-ingress-go/stage"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

var (
	sess     *session.Session
	uploader *s3manager.Uploader
	client   *s3.S3
)

func init() {
	sess = session.Must(session.NewSession())
	uploader = s3manager.NewUploader(sess)
	client = s3.New(sess)
}

// Stage stores the file in s3 and returns a presigned url
func (s *Stager) Stage(in *stage.Input) (string, error) {
	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(in.Key),
		Body:   in.Payload,
	})
	if err != nil {
		return "", errors.New("Failed to upload to s3: " + err.Error())
	}

	return s.GetURL(in.Key)
}

// GetURL gets a Presigned URL from S3
func (s *Stager) GetURL(requestID string) (string, error) {
	req, _ := client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(requestID),
	})
	url, err := req.Presign(24 * time.Hour)
	if err != nil {
		return "", errors.New("Failed to generate persigned url: " + err.Error())
	}

	return url, nil
}

// Reject moves a payload to the rejected bucket
func (s *Stager) Reject(requestID string) error {
	return s.copy(&bucketKey{
		Bucket: s.Bucket,
		Key:    requestID,
	}, s.Rejected)
}

type bucketKey struct {
	Bucket string
	Key    string
}

func (s *Stager) copy(from *bucketKey, toBucket string) error {
	src := from.Bucket + "/" + from.Key
	input := &s3.CopyObjectInput{
		Bucket:     aws.String(toBucket),
		CopySource: aws.String(src),
		Key:        aws.String(from.Key),
	}
	_, err := client.CopyObject(input)
	if err != nil {
		return errors.New("Failed to copy from " + src + " to " + toBucket)
	}
	return nil
}
