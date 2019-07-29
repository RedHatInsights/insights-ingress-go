package inventory_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	i "github.com/redhatinsights/insights-ingress-go/interactions/inventory"
	"github.com/redhatinsights/insights-ingress-go/validators"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Inventory", func() {

	var (
		validJSON string = `{"ip_addresses": ["127.0.0.1"],
		"fqdn": "localhost.localdomain",
		"mac_addresses": ["1234-5678-abcd-efgh"],
		"insights_id": "1awekljf234b24bn",
		"subscription_manager_id": "boopboop",
		"machine_id": "1awekljf234b24bn",
		"account": "000001"}`

		badJSON string = `notatallajsondoc`

		invResponse string = `{"data": [{"status": 200,
	"host": {"id": "123456"}}]}`

		invBadResponse string = `{"error": "must include account number"}`
	)

	r := []byte(validJSON)

	Describe("Calling GetID", func() {
		It("should return a valid JSON response", func() {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(207)
				fmt.Fprintln(w, invResponse)
			}))

			defer ts.Close()
			h := &i.HTTP{Endpoint: ts.URL}
			id, err := h.GetID(validators.Metadata{}, "", "")
			Expect(err).To(BeNil())
			Expect(id).To(Equal("123456"))
		})

		It("should fail on invalid json", func() {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(415)
				fmt.Fprintln(w, invBadResponse)
			}))
			defer ts.Close()
			h := &i.HTTP{Endpoint: ts.URL}
			id, err := h.GetID(validators.Metadata{}, "", "")
			Expect(id).To(Equal(""))
			Expect(err.Error()).To(Equal(invBadResponse + "\n"))
		})
	})

	Describe("Posting to Inventory", func() {
		It("should return a valid JSON response", func() {

			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(207)
				fmt.Fprintln(w, invResponse)
			}))

			defer ts.Close()
			jd := []byte(validJSON)
			res, _ := i.Post("12345", jd, ts.URL)

			Expect(res.StatusCode).To(Equal(207))
			Expect(res.Header.Get("Content-Type")).To(Equal("application/json"))

			id, err := i.ParseResponse(res)
			Expect(err).To(BeNil())
			Expect(id).To(Equal("123456"))
		})

		It("should fail on bad JSON", func() {

			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(415)
				fmt.Fprintln(w, invBadResponse)
			}))
			defer ts.Close()
			jd := []byte(badJSON)
			res, _ := i.Post("12345", jd, ts.URL)

			Expect(res.StatusCode).To(Equal(415))
			Expect(res.Header.Get("Content-Type")).To(Equal(("application/json")))

			id, err := i.ParseResponse(res)
			Expect(id).To(Equal(""))
			Expect(err).ToNot(BeNil())
		})
	})

	Describe("Creating a post", func() {
		It("should return valid data", func() {

			var b validators.Metadata
			err := json.Unmarshal(r, &b)
			Expect(err).To(BeNil())

			var m []validators.Metadata
			var x interface{}

			data, _ := i.FormatPost(b, "000001")

			err = json.Unmarshal(data, &x)
			fmt.Printf("%s\n", x)

			err = json.Unmarshal(data, &m)

			Expect(err).To(BeNil())
			Expect(m[0].Account).To(Equal("000001"))
			Expect(m[0].IPAddresses).To(ContainElement("127.0.0.1"))

		})
	})
})
