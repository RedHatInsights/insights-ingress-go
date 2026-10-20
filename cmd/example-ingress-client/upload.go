package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
)

func buildRequestBody(f io.Reader, uploadType string, filename string) (*bytes.Buffer, string) {
	requestBody := &bytes.Buffer{}
	writer := multipart.NewWriter(requestBody)
	defer writer.Close()

	innerContentType := fmt.Sprintf("application/vnd.redhat.%s.filename+tgz", uploadType)
	contentDisposition := fmt.Sprintf(`form-data; name="file"; filename="%s"`, filename)

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", contentDisposition)
	h.Set("Content-Type", innerContentType)

	part, err := writer.CreatePart(h)
	if err != nil {
		fmt.Println("Couldn't create form-file: ", err)
		os.Exit(1)
	}

	_, err = io.Copy(part, f)
	if err != nil {
		fmt.Println("Failed to copy contents to file: ", err)
		os.Exit(1)
	}

	outerContentType := fmt.Sprintf("multipart/form-data; boundary=%s", writer.Boundary())

	return requestBody, outerContentType
}

func postData(requestURL string, requestBody *bytes.Buffer, outerContentType string) {
	req, err := http.NewRequest(http.MethodPost, requestURL, requestBody)
	if err != nil {
		fmt.Println("Couldn't create request:", err)
		os.Exit(1)
	}

	req.Header.Add("Content-Type", outerContentType)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error making http request: ", err)
		os.Exit(1)
	}

	fmt.Printf("Response status code: %d\n", res.StatusCode)

	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Couldn't read response body: ", err)
		os.Exit(1)
	}
	fmt.Printf("response body: %s\n", resBody)
}

func main() {

	serverPort := 3000
	requestURL := fmt.Sprintf("http://localhost:%d/api/ingress/v1/upload", serverPort)

	filePath := "/etc/hosts"
	uploadType := "tasks"
	fileNameInArchive := "hosts.txt"

	fmt.Println("Reading file at:", filePath)
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Failed to read output file: ", err)
		os.Exit(1)
	}

	requestBody, outerContentType := buildRequestBody(file, uploadType, fileNameInArchive)

	postData(requestURL, requestBody, outerContentType)
}
