package inventory_test

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"

	i "github.com/redhatinsights/insights-ingress-go/interactions/inventory"

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

		emptyFields string = `{"ip_addresses": ["127.0.0.1"],
		"mac_addresses": ["1234-5678-abcd-efgh"],
		"insights_id": "1awekljf234b24bn",
		"subscription_manager_id": "boopboop",
		"account": "000001"}`

		badJSON string = `notatallajsondoc`

		invResponse string = `{"data": [{"status": 200,
	"host": {"id": "123456"}}]}`

		invBadResponse string = `{"error": "must include account number"}`
	)

	var r io.Reader
	var b io.Reader
	var bj io.Reader
	var p io.ReadCloser
	var bp io.ReadCloser

	r = strings.NewReader(validJSON)
	b = strings.NewReader(emptyFields)
	bj = strings.NewReader(badJSON)
	p = ioutil.NopCloser(bytes.NewReader([]byte(invResponse)))
	bp = ioutil.NopCloser(bytes.NewReader([]byte(badJSON)))

	Describe("Submitting JSON data to inventory", func() {
		It("should return a valid metadata object", func() {
			response, err := i.GetJSON(r)
			Expect(response.Account).To(Equal("000001"))
			Expect(response.IPAddresses).To(ContainElement("127.0.0.1"))
			Expect(response.FQDN).To(Equal("localhost.localdomain"))
			Expect(response.InsightsID).To(Equal("1awekljf234b24bn"))
			Expect(response.MachineID).To(Equal("1awekljf234b24bn"))
			Expect(response.SubManID).To(Equal("boopboop"))
			Expect(response.MacAddresses).To(ContainElement("1234-5678-abcd-efgh"))
			Expect(err).To(BeNil())
		})

		It("should handle empty elements", func() {
			response, err := i.GetJSON(b)
			Expect(response.FQDN).To(BeEmpty())
			Expect(response.MachineID).To(BeEmpty())
			Expect(response.IPAddresses).To(ContainElement("127.0.0.1"))
			Expect(err).To(BeNil())
		})

		It("should error on bad JSON", func() {
			_, err := i.GetJSON(bj)
			Expect(err).NotTo(BeNil())
		})
	})

	Describe("Receiving JSON response from inventory", func() {
		It("should contain a host ID", func() {
			response, err := i.FormatJSON(p)
			Expect(response.Data[0].Host.ID).To(Equal("123456"))
			Expect(response.Data[0].Status).To(Equal(200))
			Expect(err).To(BeNil())
		})

		It("should fail on bad JSON from inventory", func() {
			_, err := i.FormatJSON(bp)
			Expect(err).NotTo(BeNil())
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
			res, _ := i.Post("12345", r, ts.URL)

			Expect(res.StatusCode).To(Equal(207))
			Expect(res.Header.Get("Content-Type")).To(Equal("application/json"))
		})

		It("should fail on bad JSON", func() {

			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(415)
				fmt.Fprintln(w, invBadResponse)
			}))
			defer ts.Close()

			res, _ := i.Post("12345", r, ts.URL)

			Expect(res.StatusCode).To(Equal(415))
			Expect(res.Header.Get("Content-Type")).To(Equal(("application/json")))
		})
	})
})
