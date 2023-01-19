package upload

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httputil"
	"regexp"
	"sort"
	"strconv"
	"time"

	"github.com/redhatinsights/insights-ingress-go/internal/announcers"
	"github.com/redhatinsights/insights-ingress-go/internal/config"
	l "github.com/redhatinsights/insights-ingress-go/internal/logger"
	"github.com/redhatinsights/insights-ingress-go/internal/stage"
	"github.com/redhatinsights/insights-ingress-go/internal/validators"
	"github.com/redhatinsights/platform-go-middlewares/identity"
	"github.com/redhatinsights/platform-go-middlewares/request_id"
	"github.com/sirupsen/logrus"
)

type responseBody struct {
	RequestID string     `json:"request_id"`
	Upload    uploadData `json:"upload,omitempty"`
}

type uploadData struct {
	Account string `json:"account_number,omitempty"`
	OrgID   string `json:"org_id,omitempty"`
}

// GetFile verifies that the proper upload field is in place and returns the file
func GetFile(r *http.Request, m int64) (multipart.File, *multipart.FileHeader, error) {
	// Reduce memory usage of multipart form. This indicates of memory will be used
	// before writting to disk. Default: 8MB
	err := r.ParseMultipartForm(m)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to parse form data: %v", err)
	}

	file, fileHeader, fileErr := r.FormFile("file")
	if fileErr == nil {
		return file, fileHeader, nil
	}
	file, fileHeader, uploadErr := r.FormFile("upload")
	if uploadErr == nil {
		return file, fileHeader, nil
	}
	keys := make([]string, 0, len(r.PostForm))
	for name := range r.PostForm {
		keys = append(keys, name)
	}
	sort.Strings(keys)
	return nil, nil, fmt.Errorf("Unable to find file (%v) or upload (%v) parts in %v", fileErr, uploadErr, keys)
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
	md.Reporter = "ingress"
	md.StaleTimestamp = time.Now().AddDate(0, 0, 30)
	return &md, nil
}

// isTestRequest allows for two different test types from clients
// Current clients test using form data to the upload endpoint
// Legacy and satellite clients send a message body json of {"test": "test"}
// This function is meant to allow for both tests and use regex in the event that the
// json is sent differently in the message body depending on client version
func isTestRequest(r *http.Request) bool {
	r.ParseForm()
	if r.FormValue("test") == "test" {
		return true
	}

	if r.Header.Get("Content-Type") == "application/json" {
		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)
		body := buf.String()
		matched, _ := regexp.Match(`\{\s*\"test\"\s*\:\s*\"test\"\s*\}`, []byte(body))
		if matched {
			return true
		}
	}

	return false
}

func createUploadResponse(vr *validators.Request) ([]byte, error) {
	upload := uploadData{Account: vr.Account, OrgID: vr.OrgID}
	response := responseBody{RequestID: vr.RequestID, Upload: upload}
	jsonBody, err := json.Marshal(response)

	return jsonBody, err
}

func handleTestRequest(reqID string, identity identity.Identity, w http.ResponseWriter, logerr func(msg string, err error)) {
	mockVr := &validators.Request{
		RequestID: reqID,
		Account:   identity.AccountNumber,
		OrgID:     identity.OrgID,
	}
	w.WriteHeader(http.StatusOK)
	jsonBody, err := createUploadResponse(mockVr)
	if err != nil {
		logerr("Unable to marshal JSON response body", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(jsonBody)
}

// NewHandler returns a http handler configured with a Pipeline
func NewHandler(
	stager stage.Stager,
	validator validators.Validator,
	tracker announcers.Announcer,
	cfg config.IngressConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var id identity.XRHID
		userAgent := r.Header.Get("User-Agent")
		reqID := request_id.GetReqID(r.Context())
		requestLogger := l.Log.WithFields(logrus.Fields{"request_id": reqID, "source_host": cfg.Hostname, "name": "ingress"})

		logerr := func(msg string, err error) {
			requestLogger.WithFields(logrus.Fields{"error": err}).Error(msg)
		}

		if cfg.Auth == true {
			id = identity.Get(r.Context())
		}

		if cfg.Debug && cfg.DebugUserAgent.MatchString(userAgent) {
			dumpBytes, err := httputil.DumpRequest(r, true)
			if err != nil {
				logerr("debug: failed to dump request", err)
			} else {
				requestLogger.WithFields(logrus.Fields{"raw_request": dumpBytes}).Info("dumping request")
			}
		}

		if isTestRequest(r) {
			handleTestRequest(reqID, id.Identity, w, logerr)
			return
		}

		if id.Identity.OrgID == "" && id.Identity.Internal.OrgID != "" {
			id.Identity.OrgID = id.Identity.Internal.OrgID
		}

		if denyRequestBasedOnIdentity(cfg, id) {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("Upload denied. Please contact Red Hat Support."))
			return
		}

		incRequests(userAgent)
		file, fileHeader, err := GetFile(r, cfg.MaxUploadMem)
		if err != nil {
			errString := "File or upload field not found"
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(errString))
			requestLogger.WithFields(logrus.Fields{"request_id": reqID, "status_code": http.StatusBadRequest, "account": id.Identity.AccountNumber, "org_id": id.Identity.OrgID}).Info(errString)
			logerr("Invalid upload payload", err)
			return
		}
		// If we exit early we need to make sure this gets closed
		// later we will close this via the stageInput.close()
		// in that case, this defer will return an error because
		// the file is already closed.
		defer file.Close()
		contentType := fileHeader.Header.Get("Content-Type")
		size := fileHeader.Size

		observeSize(userAgent, size)

		requestLogger = requestLogger.WithFields(logrus.Fields{"content-type": contentType, "size": size, "request_id": reqID, "account": id.Identity.AccountNumber, "org_id": id.Identity.OrgID})

		requestLogger.Debug("ContentType received from client")
		serviceDescriptor, validationErr := getServiceDescriptor(contentType)
		if validationErr != nil {
			logerr("Unable to validate", validationErr)
			w.WriteHeader(http.StatusUnsupportedMediaType)
			requestLogger.WithFields(logrus.Fields{"status_code": http.StatusUnsupportedMediaType}).Info("Unable to validate")
			return
		}

		if err := validator.ValidateService(serviceDescriptor); err != nil {
			requestLogger.WithFields(logrus.Fields{"status_code": http.StatusUnsupportedMediaType}).Info("Unrecognized Service")
			logerr("Unrecognized service", err)
			w.WriteHeader(http.StatusUnsupportedMediaType)
			return
		}

		b64Identity := r.Header.Get("x-rh-identity")

		vr := &validators.Request{
			RequestID:   reqID,
			OrgID:       id.Identity.OrgID,
			Size:        fileHeader.Size,
			Service:     serviceDescriptor.Service,
			Category:    serviceDescriptor.Category,
			B64Identity: b64Identity,
			ContentType: contentType,
		}

		var exceedsSizeLimit bool

		if val, ok := cfg.MaxSizeMap[vr.Service]; ok {
			fileSize, _ := strconv.ParseInt(val, 10, 64)
			if fileHeader.Size > fileSize {
				exceedsSizeLimit = true
			}
		} else if fileHeader.Size > cfg.DefaultMaxSize {
			exceedsSizeLimit = true
		} else {
			exceedsSizeLimit = false
		}

		if exceedsSizeLimit {
			requestLogger.WithFields(logrus.Fields{"status_code": http.StatusRequestEntityTooLarge}).Info(fmt.Sprintf("File exceeds maximum file size (%v) for upload: %v", cfg.DefaultMaxSize, fileHeader.Size))
			w.WriteHeader(http.StatusRequestEntityTooLarge)
			w.Write([]byte(fmt.Sprintf("File exceeds maximum file size (%v) for upload: %v", cfg.DefaultMaxSize, fileHeader.Size)))
			return
		}

		if cfg.Auth == true {
			vr.Account = id.Identity.AccountNumber
			vr.Principal = id.Identity.OrgID
			vr.OrgID = id.Identity.OrgID
			requestLogger = requestLogger.WithFields(logrus.Fields{"account": vr.Account, "orgid": vr.OrgID})
		}

		md, err := GetMetadata(r)
		if err != nil {
			requestLogger.WithFields(logrus.Fields{"error": err}).Debug("Failed to read metadata")
		} else {
			vr.Metadata = *md
		}

		ps := &announcers.Status{
			Account:   vr.Account,
			OrgID:     vr.OrgID,
			RequestID: reqID,
			Status:    "received",
			StatusMsg: "Payload received by ingress",
		}
		requestLogger.Info("Payload received")
		tracker.Status(ps)

		stageInput := &stage.Input{
			Payload: file,
			Key:     reqID,
			Account: vr.Account,
			OrgId:   vr.OrgID,
			Size:    size,
		}

		start := time.Now()
		url, err := stager.Stage(stageInput)
		stageInput.Close()
		observeStageElapsed(time.Since(start))
		if err != nil {
			logerr("Error staging", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		vr.URL = url
		vr.Timestamp = time.Now()

		ps = &announcers.Status{
			Account:   vr.Account,
			OrgID:     vr.OrgID,
			RequestID: vr.RequestID,
			Status:    "success",
			StatusMsg: fmt.Sprintf("Sent to validation service: %s", vr.Service),
		}
		requestLogger.WithFields(logrus.Fields{"service": vr.Service}).Info("Payload sent to validation service")
		tracker.Status(ps)

		validator.Validate(vr)

		jsonBody, err := createUploadResponse(vr)
		if err != nil {
			logerr("Unable to marshal JSON response body", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		metadata, err := readMetadataPart(r)
		if vr.Service == "advisor" && metadata == nil {
			w.WriteHeader(http.StatusCreated)
		} else {
			w.WriteHeader(http.StatusAccepted)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonBody)
	}
}

func denyRequestBasedOnIdentity(cfg config.IngressConfig, id identity.XRHID) bool {

	for _, orgID := range cfg.DenyList {
		if orgID == id.Identity.OrgID {
			return true
		}
	}

	return false
}
