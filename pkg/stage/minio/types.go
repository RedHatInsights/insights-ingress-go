package minio

import "github.com/minio/minio-go/v6"

// Stager provides the mechanism to stage a payload via aws S3
type Stager struct {
	Bucket string
	Client *minio.Client
}
