package inventory

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/redhatinsights/insights-ingress-go/validators"
)

// Post JSON data to given URL
func Post(identity string, data []byte, url string) (*http.Response, error) {

	client := &http.Client{
		Timeout: time.Second * 10,
	}

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(data))
	req.Header.Add("x-rh-identity", identity)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// HTTP enables configuration of connection settings
type HTTP struct {
	Endpoint string
}

// FormatPost returns data in the format that Inventory accepts
func FormatPost(metadata validators.Metadata, account string) ([]byte, error) {
	metadata.Account = account
	metadata.Reporter = "ingress"
	return json.Marshal([]validators.Metadata{metadata})
}

// ParseResponse parses the inventory_id from an Inventory response or returns the error
func ParseResponse(resp *http.Response) (string, error) {
	r := &Response{}
	if resp.StatusCode == 207 {
		err := json.NewDecoder(resp.Body).Decode(&r)
		if err != nil {
			return "", errors.New("Failed to unmarshal inventory response")
		}
		return r.Data[0].Host.ID, nil
	}

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	bodyString := string(bodyBytes)
	return "", errors.New(bodyString)
}

// GetID does an HTTP Request with the metadata provided
func (h *HTTP) GetID(metadata validators.Metadata, account string, ident string) (string, error) {

	data, err := FormatPost(metadata, account)
	if err != nil {
		return "", err
	}

	resp, err := Post(ident, data, h.Endpoint)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	return ParseResponse(resp)
}
