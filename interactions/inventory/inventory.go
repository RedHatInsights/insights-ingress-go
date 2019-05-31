package inventory

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

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
	}

	return dj, nil
}

// FormatJSON encodes the inventory response
func FormatJSON(response io.ReadCloser) Inventory {

	var r Inventory

	body, err := ioutil.ReadAll(response)
	if err != nil {
		l.Log.Error("Unable to read inventory response", zap.Error(err))
	}
	err = json.Unmarshal(body, &r)
	if err != nil {
		l.Log.Error("Unable to unmarshal inventory JSON Response", zap.Error(err))
	}

	return r
}

// PostInventory does an HTTP Request with the metadata provided
func PostInventory(vr *validators.Request) (string, error) {

	var r Inventory

	cfg := config.Get()
	postBody, err := GetJSON(vr.Metadata)
	if err != nil {
		l.Log.Error("Unable to get valid PostBody", zap.Error(err),
			zap.String("request_id", vr.RequestID))
		return "", err
	}

	postBody.Account = vr.Account

	post, _ := json.Marshal([]Metadata{*&postBody})

	client := &http.Client{}
	req, _ := http.NewRequest("POST", cfg.InventoryURL, bytes.NewReader(post))
	req.Header.Add("x-rh-identity", vr.B64Identity)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		l.Log.Error("Unable to contact Inventory", zap.Error(err),
			zap.String("request_id", vr.RequestID))
	}
	if resp.StatusCode == 207 {
		r = FormatJSON(resp.Body)
	} else if resp.StatusCode == 415 {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		bodyString := string(bodyBytes)
		l.Log.Error("Inventory post failure", zap.String("error", bodyString),
			zap.String("request_id", vr.RequestID))
	}
	l.Log.Info("Successfully post to Inventory", zap.String("request_id", vr.RequestID))
	return r.Data[0].Host.ID, nil
}
