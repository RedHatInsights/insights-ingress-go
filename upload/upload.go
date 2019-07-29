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
	"github.com/redhatinsights/insights-ingress-go/pipeline"
	"github.com/redhatinsights/insights-ingress-go/stage"
	"github.com/redhatinsights/insights-ingress-go/validators"
	"github.com/redhatinsights/platform-go-middlewares/identity"
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

func readMetadataPart(r *http.Request) ([]byte, error) {
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

// GetMetadata returns metadata content from a file or value part
func GetMetadata(r *http.Request) (*validators.Metadata, error) {
	part, err := readMetadataPart(r)
	if err != nil {
		return nil, err
	}
	var md validators.Metadata
	if err = json.Unmarshal(part, &md); err != nil {
		return nil, err
	}
	return &md, nil
}

// NewHandler returns a http handler configured with a Pipeline
func NewHandler(p *pipeline.Pipeline) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userAgent := r.Header.Get("User-Agent")
		reqID := request_id.GetReqID(r.Context())
		logReqID := zap.String("request_id", reqID)

		logerr := func(msg string, err error) {
			l.Log.Error(msg, zap.Error(err), logReqID)
		}

		incRequests(userAgent)
		file, fileHeader, err := GetFile(r)
		if err != nil {
			w.WriteHeader(http.StatusUnsupportedMediaType)
			logerr("Unable to find `file` or `upload` parts", err)
			return
		}
		observeSize(userAgent, fileHeader.Size)

		serviceDescriptor, validationErr := getServiceDescriptor(fileHeader.Header.Get("Content-Type"))
		if validationErr != nil {
			logerr("Unable to validate", validationErr)
			w.WriteHeader(http.StatusUnsupportedMediaType)
			return
		}

		if fileHeader.Size > config.Get().MaxSize {
			l.Log.Info("File exceeds maximum file size for upload", zap.Int64("size", fileHeader.Size), zap.String("request_id", reqID))
			w.WriteHeader(413)
			return
		}

		if err := p.Validator.ValidateService(serviceDescriptor); err != nil {
			logerr("Unrecognized service", err)
			w.WriteHeader(http.StatusUnsupportedMediaType)
			return
		}

		b64Identity := r.Header.Get("x-rh-identity")

		vr := &validators.Request{
			RequestID:   reqID,
			Size:        fileHeader.Size,
			Service:     serviceDescriptor.Service,
			Category:    serviceDescriptor.Category,
			B64Identity: b64Identity,
		}

		if config.Get().Auth == true {
			id := identity.Get(r.Context())
			vr.Account = id.Identity.AccountNumber
			vr.Principal = id.Identity.Internal.OrgID
		}

		md, err := GetMetadata(r)
		if err != nil {
			l.Log.Debug("Failed to read metadata", zap.Error(err), logReqID)
		} else {
			vr.Metadata = *md
			vr.ID, err = p.Inventory.GetID(*md, vr.Account, b64Identity)
			if err != nil {
				logerr("Failed to post to inventory", err)
			} else {
				l.Log.Info("Successfully posted to inventory", logReqID, zap.String("inventory_id", vr.ID))
			}
		}

		ps := &validators.Status{
			Account:   vr.Account,
			Service:   "ingress",
			RequestID: reqID,
			Status:    "received",
			StatusMsg: "Payload recived by ingress",
			Date:      time.Now().UTC(),
		}
		l.Log.Info("Payload received", logReqID)
		p.Tracker.Status(ps)

		stageInput := &stage.Input{
			Payload: file,
			Key:     reqID,
		}

		start := time.Now()
		url, err := p.Stager.Stage(stageInput)
		stageInput.Close()
		observeStageElapsed(time.Since(start))
		if err != nil {
			logerr("Error staging", err)
			return
		}

		vr.URL = url
		vr.Timestamp = time.Now()

		ps = &validators.Status{
			Account:   vr.Account,
			Service:   "ingress",
			RequestID: vr.RequestID,
			Status:    "processing",
			StatusMsg: "Sent to validation service",
			Date:      time.Now().UTC(),
		}
		l.Log.Info("Payload sent to validation service", logReqID)
		p.Tracker.Status(ps)

		p.Validator.Validate(vr)

		w.WriteHeader(http.StatusAccepted)
	}
}
