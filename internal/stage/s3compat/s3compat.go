package s3compat

import (
	"errors"
	"time"
	"fmt"

	"github.com/minio/minio-go/v6"
	"github.com/redhatinsights/insights-ingress-go/internal/config"
	"github.com/redhatinsights/insights-ingress-go/internal/stage"
)

// Stager provides the mechanism to stage a payload via aws S3
type Stager struct {
	Bucket string
	Client *minio.Client
}

// GetClient gets the s3 compatible client info
func GetClient(cfg *config.IngressConfig, stager *Stager) stage.Stager {
	var endpoint string
	storageCfg := cfg.StorageConfig
	if storageCfg.StorageEndpoint == "" {
		endpoint = "s3.amazonaws.com"
	} else {
		endpoint = storageCfg.StorageEndpoint
	}
	accessKeyID := storageCfg.StorageAccessKey
	secretAccessKey := storageCfg.StorageSecretKey
	useSSL := storageCfg.UseSSL

	fmt.Print(storageCfg)

	if storageCfg.StorageRegion != "" { 
		stager.Client, _ = minio.NewWithRegion(endpoint, accessKeyID, secretAccessKey, useSSL, storageCfg.StorageRegion)
	} else {
		stager.Client, _ = minio.New(endpoint, accessKeyID, secretAccessKey, useSSL)
	}

	return stager
}

// Stage stores the file in s3 compatible storage and returns a presigned url
func (s *Stager) Stage(in *stage.Input) (string, error) {
	bucketName := s.Bucket
	objectName := in.Key
	object := in.Payload
	contentType := "application/gzip"

	_, err := s.Client.PutObject(bucketName,
		objectName,
		object,
		in.Size,
		minio.PutObjectOptions{
			ContentType: contentType,
			UserMetadata: map[string]string{
				"requestID": in.Key,
				"account":   in.Account,
				"org":       in.OrgId,
			},
		},
	)
	if err != nil {
		return "", errors.New("Failed to upload to storage" + err.Error())
	}
	return s.GetURL(in.Key)
}

// GetURL retrieves a presigned url from s3 compatible storage
func (s *Stager) GetURL(requestID string) (string, error) {
	url, err := s.Client.PresignedGetObject(s.Bucket, requestID, time.Second*24*60*60, nil)
	if err != nil {
		return "", errors.New("Failed to generate presigned url: " + err.Error())
	}
	return url.String(), nil
}
