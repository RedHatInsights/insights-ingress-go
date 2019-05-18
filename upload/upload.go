package upload

import (
	"errors"
	"log"
	"net/http"
	"regexp"

	"cloud.redhat.com/ingress/pipeline"
	"cloud.redhat.com/ingress/stage"
	"github.com/go-chi/chi/middleware"
	"github.com/redhatinsights/platform-go-middlewares/identity"
)

var contentTypePat = regexp.MustCompile(`application/vnd\.redhat\.(\w+)\.(\w+)`)

func validate(contentType string) (*TopicDescriptor, error) {
	// look the content type up in a static map
	// else parse it
	m := contentTypePat.FindStringSubmatch(contentType)
	if m == nil {
		return nil, errors.New("Failed to match on Content-Type: " + contentType)
	}
	return &TopicDescriptor{
		Service:  m[1],
		Category: m[2],
	}, nil
}

// NewHandler returns a http handler configured with a Stager
func NewHandler(p *pipeline.Pipeline) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// look for `file` part
		file, fileHeader, err := r.FormFile("file")
		if err != nil {
			log.Printf("Did not find `file` part: %v", err)
			w.WriteHeader(http.StatusUnsupportedMediaType)
			return
		}

		topicDescriptor, validationErr := validate(fileHeader.Header.Get("Content-Type"))
		if validationErr != nil {
			log.Printf("Did not validate: %v", validationErr)
			w.WriteHeader(http.StatusUnsupportedMediaType)
			return
		}

		reqID := middleware.GetReqID(r.Context())

		stageInput := &stage.Input{
			Reader: file,
			Key:    reqID,
		}

		// look for the metadata part
		metadata, metadataHeader, err := r.FormFile("metadata")
		if err != nil {
			log.Printf("Did not find `metadata` part: %v", err)
		} else {
			log.Printf("%v, %v", metadata, metadataHeader)
			stageInput.Metadata = metadata
		}

		id, err := identity.Get(r.Context())
		if err != nil {
			log.Printf("Failed to fetch identity from context")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		vr := &pipeline.ValidationRequest{
			Account:   id.AccountNumber,
			Principal: id.Internal.OrgID,
			PayloadID: reqID,
			Size:      fileHeader.Size,
			Service:   topicDescriptor.Service,
			Category:  topicDescriptor.Category,
			Metadata:  metadata,
		}

		// copy to s3
		go p.Submit(stageInput, vr)
		// broadcast on kafka topic
		// return accepted response
		w.Header().Set("X-Request-Id", reqID)
		w.WriteHeader(http.StatusAccepted)
	}
}
