package inventory

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/redhatinsights/insights-ingress-go/config"
	l "github.com/redhatinsights/insights-ingress-go/logger"
	"github.com/redhatinsights/insights-ingress-go/validators"
	"go.uber.org/zap"
)

// GetJSON decodes the incoming metadata from an upload
func GetJSON(metadata io.Reader) (Metadata, error) {

	var dj Metadata
	err := json.NewDecoder(metadata).Decode(&dj)
	if err != nil {
		l.Log.Error("Unable to decode metadata JSON", zap.Error(err))
		return Metadata{}, err
	}

	return dj, nil
}

// FormatJSON encodes the inventory response
func FormatJSON(response io.ReadCloser) (Inventory, error) {

	var r Inventory

	body, err := ioutil.ReadAll(response)
	if err != nil {
		l.Log.Error("Unable to read inventory response", zap.Error(err))
		return Inventory{}, err
	}
	err = json.Unmarshal(body, &r)
	if err != nil {
		l.Log.Error("Unable to unmarshal inventory JSON Response", zap.Error(err))
		return Inventory{}, err
	}

	return r, nil
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

	return resp, err
}

// CreatePost puts together the post to be sent to inventory
func CreatePost(vr *validators.Request) ([]byte, error) {

	postBody, err := GetJSON(vr.Metadata)
	if err != nil {
		l.Log.Error("Unable to get valid PostBody", zap.Error(err),
			zap.String("request_id", vr.RequestID))
		return nil, err
	}

	postBody.Account = vr.Account

	post, _ := json.Marshal([]Metadata{*&postBody})

	return post, nil
}

// CallInventory does an HTTP Request with the metadata provided
func CallInventory(vr *validators.Request) (string, error) {

	var r Inventory

	cfg := config.Get()
	data, err := CreatePost(vr)

	resp, err := Post(vr.B64Identity, data, cfg.InventoryURL)
	if err != nil {
		l.Log.Error("Unable to post to Inventory", zap.Error(err),
			zap.String("request_id", vr.RequestID))
	}
	if resp.StatusCode == 207 {
		r, err = FormatJSON(resp.Body)
		l.Log.Info("Successfully post to Inventory", zap.String("request_id", vr.RequestID))
	} else {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		bodyString := string(bodyBytes)
		l.Log.Error("Inventory post failure", zap.String("error", bodyString),
			zap.String("request_id", vr.RequestID))
	}
	return r.Data[0].Host.ID, nil
}
