package upload

import (
	"errors"
	"log"
	"net/http"
	"regexp"

	"github.com/go-chi/chi/middleware"
	"github.com/redhatinsights/insights-ingress-go/config"
	"github.com/redhatinsights/insights-ingress-go/pipeline"
	"github.com/redhatinsights/insights-ingress-go/stage"
	"github.com/redhatinsights/insights-ingress-go/validators"
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

// NewHandler returns a http handler configured with a Pipeline
func NewHandler(p *pipeline.Pipeline) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		b64Identity := r.Header.Get("x-rh-identity")

		reqID := middleware.GetReqID(r.Context())

		stageInput := &stage.Input{
			Reader: file,
			Key:    reqID,
		}

		metadata, metadataHeader, err := r.FormFile("metadata")
		if err != nil {
			log.Printf("Did not find `metadata` part: %v", err)
		} else {
			log.Printf("%v, %v", metadata, metadataHeader)
			stageInput.Metadata = metadata
		}

		vr := &validators.Request{
			RequestID:   reqID,
			Size:        fileHeader.Size,
			Service:     topicDescriptor.Service,
			Category:    topicDescriptor.Category,
			B64Identity: b64Identity,
			Metadata:    metadata,
		}

		if config.Get().Auth == true {
			id := identity.Get(r.Context())
			vr.Account = id.Identity.AccountNumber
			vr.Principal = id.Identity.Internal.OrgID
		}

		go p.Submit(stageInput, vr)

		w.Header().Set("X-Request-Id", reqID)
		w.WriteHeader(http.StatusAccepted)
	}
}
