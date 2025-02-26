package filebased

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/redhatinsights/insights-ingress-go/internal/stage"
)

// Stager provides the mechanism to stage a payload to the file system
type FileBasedStager struct {
	StagePath string
	BaseURL   string
}

func FilterAlphanumeric(s string) string {
	var sb strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

func GetFileStorageName(requestID string) (string, error) {
	key := FilterAlphanumeric(requestID)
	if len(key) == 0 {
		return "", errors.New("bad request id format")
	}
	fileName := key + ".tar.gz"
	return fileName, nil
}

// Stage stores the file in filesystem storage and returns a presigned url
func (s *FileBasedStager) Stage(in *stage.Input) (string, error) {
	fileName, err := GetFileStorageName(in.Key)
	if err != nil {
		return "", err
	}
	file := in.Payload
	filePath := filepath.Join(s.StagePath, fileName)
	dst, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		return "", err
	}
	return s.GetURL(in.Key)
}

// GetURL retrieves a presigned url from filesystem storage
func (s *FileBasedStager) GetURL(requestID string) (string, error) {
	fileURL := s.BaseURL + "/download/" + requestID
	return fileURL, nil
}
