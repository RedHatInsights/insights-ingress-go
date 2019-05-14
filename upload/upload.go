package upload

import (
	"io"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

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

// Handle accepts incoming payloads
func Handle(w http.ResponseWriter, r *http.Request) {
	// look for `file` part
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		log.Printf("Did not find `file` part: %v", err)
		return
	}

	log.Printf("%v, %v", file, fileHeader)

	// look for the metadata part
	metadata, metadataHeader, err := r.FormFile("metadata")
	if err != nil {
		log.Printf("Did not find `metadata` part: %v", err)
	}
	log.Printf("%v, %v", metadata, metadataHeader)
	// copy to s3
	go store(file)
	// broadcast on kafka topic
	// return accepted response
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusAccepted)
	return
}
