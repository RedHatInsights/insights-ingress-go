package api_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/redhatinsights/insights-ingress-go/internal/api"
)

var _ = Describe("OpenAPI Spec", func() {
	var spec map[string]interface{}

	BeforeEach(func() {
		Expect(json.Unmarshal(api.ApiSpec, &spec)).To(Succeed())
	})

	Describe("Metadata schema for /upload", func() {
		var metadataSchema map[string]interface{}

		BeforeEach(func() {
			paths := spec["paths"].(map[string]interface{})
			upload := paths["/upload"].(map[string]interface{})
			post := upload["post"].(map[string]interface{})
			reqBody := post["requestBody"].(map[string]interface{})
			content := reqBody["content"].(map[string]interface{})
			multipart := content["multipart/form-data"].(map[string]interface{})
			schema := multipart["schema"].(map[string]interface{})
			props := schema["properties"].(map[string]interface{})
			metadataSchema = props["metadata"].(map[string]interface{})
		})

		It("should reference the Metadata component schema", func() {
			ref, hasRef := metadataSchema["$ref"]
			Expect(hasRef).To(BeTrue(), "metadata should use a $ref to the Metadata schema")
			Expect(ref).To(Equal("#/components/schemas/Metadata"))
		})
	})

	Describe("Metadata component schema", func() {
		var metadataProps map[string]interface{}

		BeforeEach(func() {
			components := spec["components"].(map[string]interface{})
			schemas := components["schemas"].(map[string]interface{})
			metadata, ok := schemas["Metadata"].(map[string]interface{})
			Expect(ok).To(BeTrue(), "Metadata schema must exist in components/schemas")
			metadataProps = metadata["properties"].(map[string]interface{})
		})

		It("should define ip_addresses as an array of strings", func() {
			field := metadataProps["ip_addresses"].(map[string]interface{})
			Expect(field["type"]).To(Equal("array"))
			items := field["items"].(map[string]interface{})
			Expect(items["type"]).To(Equal("string"))
		})

		It("should define mac_addresses as an array of strings", func() {
			field := metadataProps["mac_addresses"].(map[string]interface{})
			Expect(field["type"]).To(Equal("array"))
			items := field["items"].(map[string]interface{})
			Expect(items["type"]).To(Equal("string"))
		})

		It("should define string fields correctly", func() {
			stringFields := []string{
				"account", "org_id", "insights_id", "machine_id",
				"subscription_manager_id", "fqdn", "bios_uuid",
				"display_name", "ansible_host", "reporter", "queue_key",
			}
			for _, name := range stringFields {
				field, ok := metadataProps[name].(map[string]interface{})
				Expect(ok).To(BeTrue(), "field %s should exist", name)
				Expect(field["type"]).To(Equal("string"), "field %s should be type string", name)
			}
		})

		It("should define custom_metadata as an object with string values", func() {
			field := metadataProps["custom_metadata"].(map[string]interface{})
			Expect(field["type"]).To(Equal("object"))
			addlProps := field["additionalProperties"].(map[string]interface{})
			Expect(addlProps["type"]).To(Equal("string"))
		})

		It("should define stale_timestamp as a date-time string", func() {
			field := metadataProps["stale_timestamp"].(map[string]interface{})
			Expect(field["type"]).To(Equal("string"))
			Expect(field["format"]).To(Equal("date-time"))
		})
	})
})
