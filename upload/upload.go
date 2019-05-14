package upload

import (
	"errors"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

var contentTypePat = regexp.MustCompile(`application/vnd\.redhat\.(\w+)\.(\w+)`)

func store(file io.Reader) {
	sess := session.Must(session.NewSession())
	uploader := s3manager.NewUploader(sess)
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String("jjaggars-test"),
		Key:    aws.String("foo"),
		Body:   file,
	})
	if err != nil {
		log.Printf("Failed to upload to s3: %v", err)
		return
	}
	log.Printf("Successfully uploaded to %s.", result.Location)
}

func validate(header *multipart.FileHeader) error {
	contentType := header.Header.Get("Content-Type")
	// look the content type up in a static map
	// else parse it
	m := contentTypePat.FindStringSubmatch(contentType)
	if m == nil {
		return errors.New("Failed to match on Content-Type: " + contentType)
	}
	log.Printf("service = %s, category = %s", m[1], m[2])
	return nil
}

// Handle accepts incoming payloads
func Handle(w http.ResponseWriter, r *http.Request) {
	// look for `file` part
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		log.Printf("Did not find `file` part: %v", err)
		return
	}

	if validationErr := validate(fileHeader); validationErr != nil {
		log.Printf("Did not validate: %v", validationErr)
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}

	// look for the metadata part
	metadata, metadataHeader, err := r.FormFile("metadata")
	if err != nil {
		log.Printf("Did not find `metadata` part: %v", err)
	} else {
		log.Printf("%v, %v", metadata, metadataHeader)
	}
	// copy to s3
	go store(file)
	// broadcast on kafka topic
	// return accepted response
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusAccepted)
	return
}
