package upload

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/redhatinsights/insights-ingress-go/config"
	l "github.com/redhatinsights/insights-ingress-go/logger"
	identity "github.com/redhatinsights/insights-ingress-go/middleware"
	"github.com/redhatinsights/insights-ingress-go/pipeline"
	"github.com/redhatinsights/insights-ingress-go/stage"
	"github.com/redhatinsights/insights-ingress-go/validators"
	"github.com/redhatinsights/platform-go-middlewares/request_id"
	"go.uber.org/zap"
)

// GetFile verifies that the proper upload field is in place and returns the file
func GetFile(r *http.Request) (multipart.File, *multipart.FileHeader, error) {
	var err error
	file, fileHeader, err := r.FormFile("file")
	if err == nil {
		return file, fileHeader, nil
	}
	file, fileHeader, err = r.FormFile("upload")
	if err == nil {
		return file, fileHeader, nil
	}
	return nil, nil, err
}

// GetMetadata returns metadata content from a file or value part
func GetMetadata(r *http.Request) ([]byte, error) {
	mdf, _, err := r.FormFile("metadata")
	if err == nil {
		defer mdf.Close()
		return ioutil.ReadAll(mdf)
	}
	metadata := r.FormValue("metadata")
	if metadata != "" {
		return []byte(metadata), nil
	}

	return nil, errors.New("Failed to find metadata as a file or value")
}

// NewHandler returns a http handler configured with a Pipeline
func NewHandler(p *pipeline.Pipeline) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userAgent := r.Header.Get("User-Agent")
		reqID := request_id.GetReqID(r.Context())

		incRequests(userAgent)
		file, fileHeader, err := GetFile(r)
		if err != nil {
			w.WriteHeader(http.StatusUnsupportedMediaType)
			l.Log.Error("Unable to find `file` or `upload` parts", zap.Error(err), zap.String("request_id", reqID))
			return
		}
		observeSize(userAgent, fileHeader.Size)

		serviceDescriptor, validationErr := getServiceDescriptor(fileHeader.Header.Get("Content-Type"))
		if validationErr != nil {
			l.Log.Info("Did not validate", zap.Error(validationErr), zap.String("request_id", reqID))
			w.WriteHeader(http.StatusUnsupportedMediaType)
			return
		}

		if err := p.Validator.ValidateService(serviceDescriptor); err != nil {
			l.Log.Info("Unrecognized service", zap.Error(err), zap.String("request_id", reqID))
			w.WriteHeader(http.StatusUnsupportedMediaType)
			return
		}

		b64Identity := r.Header.Get("x-rh-identity")

		stageInput := &stage.Input{
			Payload: file,
			Key:     reqID,
		}

		metadata, err := GetMetadata(r)
		if err != nil {
			l.Log.Debug("Empty metadata", zap.Error(err), zap.String("request_id", reqID))
		}
		var md validators.Metadata
		if err = json.Unmarshal(metadata, &md); err != nil {
			l.Log.Error("Failed to unmarshal metadata", zap.Error(err), zap.String("request_id", reqID))
		}

		vr := &validators.Request{
			RequestID:   reqID,
			Size:        fileHeader.Size,
			Service:     serviceDescriptor.Service,
			Category:    serviceDescriptor.Category,
			B64Identity: b64Identity,
			Metadata:    md,
		}

		if config.Get().Auth == true {
			id := identity.Get(r.Context())
			vr.Account = id.Identity.AccountNumber
			vr.Principal = id.Identity.Internal.OrgID
		}

		if metadata != nil {
			id, err := p.Inventory.GetID(vr)
			if err != nil {
				l.Log.Error("Inventory post failure", zap.Error(err), zap.String("request_id", reqID))
			} else {
				vr.ID = id
			}
		}

		ps := &validators.Status{
			Account:   vr.Account,
			Service:   "ingress",
			RequestID: reqID,
			Status:    "recieved",
			StatusMsg: "Payload recived by ingress",
			Date:      time.Now().UTC(),
		}
		p.Tracker.Status(ps)

		go p.Submit(stageInput, vr)

		w.WriteHeader(http.StatusAccepted)
	}
}
