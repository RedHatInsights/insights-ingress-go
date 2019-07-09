package inventory

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	l "github.com/redhatinsights/insights-ingress-go/logger"
	"github.com/redhatinsights/insights-ingress-go/validators"
	"go.uber.org/zap"
)

// GetJSON decodes the incoming metadata from an upload
func GetJSON(metadata []byte) (Metadata, error) {

	var dj Metadata
	err := json.Unmarshal(metadata, &dj)
	if err != nil {
		return Metadata{}, err
	}

	return dj, nil
}

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

// CreatePost puts together the post to be sent to inventory
func CreatePost(vr *validators.Request) ([]byte, error) {

	postBody, err := GetJSON(vr.Metadata)
	if err != nil {
		return nil, err
	}

	postBody.Account = vr.Account

	post, _ := json.Marshal([]Metadata{postBody})

	return post, nil
}

// HTTP enables configuration of connection settings
type HTTP struct {
	Endpoint string
}

// GetID does an HTTP Request with the metadata provided
func (h *HTTP) GetID(vr *validators.Request) (string, error) {

	data, err := CreatePost(vr)
	if err != nil {
		return "", err
	}

	resp, err := Post(vr.B64Identity, data, h.Endpoint)
	if err != nil {
		return "", err
	}

	r := &Response{}
	if resp.StatusCode == 207 {
		l.Log.Info("Successfully post to Inventory", zap.String("request_id", vr.RequestID))
		json.NewDecoder(resp.Body).Decode(&r)
		return r.Data[0].Host.ID, nil
	}

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	bodyString := string(bodyBytes)
	return "", errors.New(bodyString)
}
