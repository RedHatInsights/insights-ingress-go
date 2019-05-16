package upload

import (
	"errors"
	"log"
	"mime/multipart"
	"net/http"
	"regexp"

	"github.com/go-chi/chi/middleware"
)

var contentTypePat = regexp.MustCompile(`application/vnd\.redhat\.(\w+)\.(\w+)`)

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

// NewHandler returns a http handler configured with a Stager
func NewHandler(stager Stager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// look for `file` part
		file, fileHeader, err := r.FormFile("file")
		if err != nil {
			log.Printf("Did not find `file` part: %v", err)
			w.WriteHeader(http.StatusUnsupportedMediaType)
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

		log.Printf("%v\n", r)
		// copy to s3
		go stager.Stage(file, middleware.GetReqID(r.Context()))
		// broadcast on kafka topic
		// return accepted response
		w.Header().Set("X-Request-Id", middleware.GetReqID(r.Context()))
		w.WriteHeader(http.StatusAccepted)
	}
}
