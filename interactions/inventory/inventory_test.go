package inventory_test

import (
	"bytes"
	"io"
	"io/ioutil"
	"strings"

	i "github.com/redhatinsights/insights-ingress-go/interactions/inventory"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Inventory", func() {

	validJSON := `{"ip_addresses": ["127.0.0.1"],
		"fqdn": "localhost.localdomain",
		"mac_addresses": ["1234-5678-abcd-efgh"],
		"insights_id": "1awekljf234b24bn",
		"subscription_manager_id": "boopboop",
		"machine_id": "1awekljf234b24bn",
		"account": "000001"}`

	emptyFields := `{"ip_addresses": ["127.0.0.1"],
		"mac_addresses": ["1234-5678-abcd-efgh"],
		"insights_id": "1awekljf234b24bn",
		"subscription_manager_id": "boopboop",
		"account": "000001"}`

	invResponse := `{"data": [{"status": 200,
	"host": {"id": "123456"}}]}`

	var r io.Reader
	var b io.Reader
	var p io.ReadCloser

	r = strings.NewReader(validJSON)
	b = strings.NewReader(emptyFields)
	p = ioutil.NopCloser(bytes.NewReader([]byte(invResponse)))

	Describe("Submitting JSON data to inventory", func() {
		It("should return a valid metadata object", func() {
			response, _ := i.GetJSON(r)
			Expect(response.Account).To(Equal("000001"))
			Expect(response.IPAddresses).To(ContainElement("127.0.0.1"))
			Expect(response.FQDN).To(Equal("localhost.localdomain"))
			Expect(response.InsightsID).To(Equal("1awekljf234b24bn"))
			Expect(response.MachineID).To(Equal("1awekljf234b24bn"))
			Expect(response.SubManID).To(Equal("boopboop"))
			Expect(response.MacAddresses).To(ContainElement("1234-5678-abcd-efgh"))
		})

		It("should handle empty elements", func() {
			response, _ := i.GetJSON(b)
			Expect(response.FQDN).To(BeEmpty())
			Expect(response.MachineID).To(BeEmpty())
			Expect(response.IPAddresses).To(ContainElement("127.0.0.1"))
		})
	})

	Describe("Receiving JSON response from inventory", func() {
		It("should contain a host ID", func() {
			response := i.FormatJSON(p)
			Expect(response.Data[0].Host.ID).To(Equal("123456"))
			Expect(response.Data[0].Status).To(Equal(200))
		})
	})
})
