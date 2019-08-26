package minio

import (
	"errors"
	"time"

	"github.com/minio/minio-go/v6"
	"github.com/redhatinsights/insights-ingress-go/config"
	"github.com/redhatinsights/insights-ingress-go/stage"
)

// GetClient gets the minio client info
func GetClient(stager *Stager) stage.Stager {
	endpoint := config.Get().MinioEndpoint
	accessKeyID := config.Get().MinioAccessKey
	secretAccessKey := config.Get().MinioSecretKey
	useSSL := false

	stager.Client, _ = minio.New(endpoint, accessKeyID, secretAccessKey, useSSL)

	return stager
}

// Stage stores the file in minio and returns a presigned url
func (s *Stager) Stage(in *stage.Input) (string, error) {
	bucketName := s.Bucket
	objectName := in.Key
	object := in.Payload
	contentType := "application/gzip"

	_, err := s.Client.PutObject(bucketName, objectName, object, -1, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return "", errors.New("Failed to upload to minio" + err.Error())
	}
	return s.GetURL(in.Key)
}

// GetURL retrieves a presigned url from minio
func (s *Stager) GetURL(requestID string) (string, error) {
	url, err := s.Client.PresignedGetObject(s.Bucket, requestID, time.Second*24*60*60, nil)
	if err != nil {
		return "", errors.New("Failed to generate presigned url: " + err.Error())
	}
	return url.String(), nil
}
