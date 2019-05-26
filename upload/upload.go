package upload

import (
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/redhatinsights/insights-ingress-go/config"
	l "github.com/redhatinsights/insights-ingress-go/logger"
	"github.com/redhatinsights/insights-ingress-go/pipeline"
	"github.com/redhatinsights/insights-ingress-go/stage"
	"github.com/redhatinsights/insights-ingress-go/validators"
	"github.com/redhatinsights/platform-go-middlewares/identity"
	"go.uber.org/zap"
)

// NewHandler returns a http handler configured with a Pipeline
func NewHandler(p *pipeline.Pipeline) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userAgent := r.Header.Get("User-Agent")

		incRequests(userAgent)

		file, fileHeader, err := r.FormFile("file")
		if err != nil {
			l.Log.Info("Did not find `file` part", zap.Error(err))
			w.WriteHeader(http.StatusUnsupportedMediaType)
			return
		}

		observeSize(userAgent, fileHeader.Size)

		serviceDescriptor, validationErr := getServiceDescriptor(fileHeader.Header.Get("Content-Type"))
		if validationErr != nil {
			l.Log.Info("Did not validate", zap.Error(validationErr))
			w.WriteHeader(http.StatusUnsupportedMediaType)
			return
		}

		if err := p.Validator.ValidateService(serviceDescriptor); err != nil {
			l.Log.Info("Unrecognized service", zap.Error(err))
			w.WriteHeader(http.StatusUnsupportedMediaType)
			return
		}

		b64Identity := r.Header.Get("x-rh-identity")

		reqID := middleware.GetReqID(r.Context())

		stageInput := &stage.Input{
			Payload: file,
			Key:     reqID,
		}

		metadata, _, err := r.FormFile("metadata")
		if err != nil {
			l.Log.Info("Did not find `metadata` part", zap.Error(err))
		}

		vr := &validators.Request{
			RequestID:   reqID,
			Size:        fileHeader.Size,
			Service:     serviceDescriptor.Service,
			Category:    serviceDescriptor.Category,
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
