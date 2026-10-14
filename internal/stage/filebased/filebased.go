package filebased

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/redhatinsights/insights-ingress-go/internal/stage"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Check verifies that the local staging directory remains available.
func (f *FileBasedStager) Check(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	probe, err := os.CreateTemp(f.StagePath, ".ingress-health-*")
	if err != nil {
		return errors.New("file staging directory unavailable")
	}
	probePath := probe.Name()
	if err := probe.Close(); err != nil {
		_ = os.Remove(probePath)
		return errors.New("file staging directory unavailable")
	}
	if err := os.Remove(probePath); err != nil {
		return errors.New("file staging directory unavailable")
	}
	return nil
}

// Stager provides the mechanism to stage a payload to the file system
type FileBasedStager struct {
	StagePath string
	BaseURL   string
}

func filterAlphanumeric(s string) string {
	var sb strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

func GetFileStorageName(requestID string, storageDir string) (string, string, error) {
	key := filterAlphanumeric(requestID)
	if len(key) == 0 {
		return "", "", errors.New("bad request id format")
	}
	fileName := key + ".tar.gz"
	filePath := filepath.Join(storageDir, fileName)
	return fileName, filePath, nil
}

// Stage stores the file in filesystem storage and returns a presigned url
func (s *FileBasedStager) Stage(ctx context.Context, in *stage.Input) (string, error) {
	_, filePath, err := GetFileStorageName(in.Key, s.StagePath)
	if err != nil {
		return "", err
	}

	_, span := otel.Tracer("ingress").Start(ctx, "stage.file.write",
		trace.WithAttributes(
			attribute.String("file.path", filePath),
			attribute.Int64("file.size", in.Size),
		))
	defer span.End()
	file := in.Payload
	dst, err := os.Create(filePath)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return "", err
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return "", err
	}
	return s.GetURL(in.Key)
}

// GetURL retrieves a presigned url from filesystem storage
func (s *FileBasedStager) GetURL(requestID string) (string, error) {
	fileURL := s.BaseURL + "/download/" + requestID
	return fileURL, nil
}
