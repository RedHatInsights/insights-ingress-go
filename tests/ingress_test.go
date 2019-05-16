package upload

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"strings"
	"testing"
	"time"

	"cloud.redhat.com/ingress/upload"
)

type FakeStager struct {
	Out chan int
}

func (s *FakeStager) Stage(file io.Reader, key string) (string, error) {
	s.Out <- 1
	fmt.Println("stager just got called")
	return "", nil
}

func makeMultipartRequest(name string, content string) (*http.Request, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition",
		fmt.Sprintf(`form-data; name="%s"; filename="%s.txt"`, name, name))
	h.Set("Content-Type", "application/vnd.redhat.unit.test")
	part, err := writer.CreatePart(h)

	if err != nil {
		return nil, err
	}

	_, err = io.Copy(part, strings.NewReader(content))
	if err != nil {
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "/upload", body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, nil
}

func TestUploadHandler(t *testing.T) {
	req, err := makeMultipartRequest("file", "testing")
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	ch := make(chan int)
	stager := &FakeStager{Out: ch}
	handler := upload.NewHandler(stager)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusAccepted {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusAccepted)
	}

	// Give the goroutine a second
	select {
	case <-ch:
		break
	case <-time.After(1 * time.Second):
		t.Errorf("stager was not called")
	}
}
