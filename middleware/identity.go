package middleware

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
)

type identityKey int

// Key the key for the XRHID in that gets added to the context
const Key identityKey = iota

// Internal is the "internal" field of an XRHID
type Internal struct {
	OrgID string `json:"org_id"`
}

type Identity struct {
	AccountNumber string   `json:"account_number"`
	Internal      Internal `json:"internal"`
}

// XRHID is the "identity" pricipal object set by Cloud Platform 3scale
type XRHID struct {
	Identity Identity `json:"identity"`
}

func getErrorText(code int, reason string) string {
	return http.StatusText(code) + ": " + reason
}

func doError(w http.ResponseWriter, code int, reason string) {
	http.Error(w, getErrorText(code, reason), code)
}

// Get returns the identity struct from the context
func Get(ctx context.Context) XRHID {
	return ctx.Value(Key).(XRHID)
}

// Identity extracts the X-Rh-Identity header and places the contents into the
// request context
func EnforceIdentity(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rawHeaders := r.Header["X-Rh-Identity"]

		// must have an x-rh-id header
		if len(rawHeaders) != 1 {
			doError(w, 400, "missing x-rh-identity header")
			return
		}

		// must be able to base64 decode header
		idRaw, err := base64.StdEncoding.DecodeString(rawHeaders[0])
		if err != nil {
			doError(w, 400, "unable to b64 decode x-rh-identity header")
			return
		}

		var jsonData XRHID
		err = json.Unmarshal(idRaw, &jsonData)
		if err != nil {
			doError(w, 400, "x-rh-identity header is does not contain valid JSON")
			return
		}

		if jsonData.Identity.AccountNumber == "" || jsonData.Identity.AccountNumber == "-1" {
			doError(w, 400, "x-rh-identity header has an invalid or missing account number")
			return
		}

		ctx := context.WithValue(r.Context(), Key, jsonData)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
