package s3compat

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/minio/minio-go/v6"
	"github.com/redhatinsights/insights-ingress-go/internal/config"
	"github.com/redhatinsights/insights-ingress-go/internal/stage"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Stager provides the mechanism to stage a payload via aws S3
type S3Stager struct {
	Bucket string
	Client *minio.Client
}

// Check verifies credentials and access to the configured bucket without
// writing customer data.
func (s *S3Stager) Check(ctx context.Context) error {
	if s.Client == nil {
		return errors.New("object storage client is not configured")
	}
	exists, err := s.Client.BucketExistsWithContext(ctx, s.Bucket)
	if err != nil {
		return fmt.Errorf("object storage bucket unavailable: %w", err)
	}
	if !exists {
		return errors.New("object storage bucket unavailable")
	}
	return nil
}

// GetClient gets the s3 compatible client info
func GetClient(cfg *config.IngressConfig, stager *S3Stager) stage.Stager {
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

	if storageCfg.StorageRegion != "" {
		stager.Client, _ = minio.NewWithRegion(endpoint, accessKeyID, secretAccessKey, useSSL, storageCfg.StorageRegion)
	} else {
		stager.Client, _ = minio.New(endpoint, accessKeyID, secretAccessKey, useSSL)
	}

	return stager
}

// Stage stores the file in s3 compatible storage and returns a presigned url
func (s *S3Stager) Stage(ctx context.Context, in *stage.Input) (string, error) {
	bucketName := s.Bucket
	objectName := in.Key
	object := in.Payload
	contentType := "application/gzip"

	_, span := otel.Tracer("ingress").Start(ctx, "S3.PutObject",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("rpc.system", "aws-api"),
			attribute.String("rpc.service", "S3"),
			attribute.String("rpc.method", "PutObject"),
			attribute.String("aws.s3.bucket", bucketName),
			attribute.String("aws.s3.key", objectName),
			attribute.Int64("file.size", in.Size),
		))
	defer span.End()

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
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return "", errors.New("Failed to upload '" + bucketName + "' to storage: " + err.Error())
	}
	return s.GetURL(in.Key)
}

// GetURL retrieves a presigned url from s3 compatible storage
func (s *S3Stager) GetURL(requestID string) (string, error) {
	url, err := s.Client.PresignedGetObject(s.Bucket, requestID, time.Second*24*60*60, nil)
	if err != nil {
		return "", errors.New("Failed to generate presigned url: " + err.Error())
	}
	return url.String(), nil
}
