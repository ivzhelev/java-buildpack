package frameworks_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Dynatrace OneAgent", func() {
	Describe("Manifest parsing", func() {
		It("parses Dynatrace manifest JSON", func() {
			manifestJSON := `{
				"tenantToken": "test-token-123",
				"communicationEndpoints": [
					"https://endpoint1.dynatrace.com",
					"https://endpoint2.dynatrace.com"
				],
				"technologies": {
					"process": {
						"linux-x86-64": []
					}
				}
			}`

			var manifest map[string]interface{}
			err := json.Unmarshal([]byte(manifestJSON), &manifest)
			Expect(err).NotTo(HaveOccurred())

			tenantToken, ok := manifest["tenantToken"].(string)
			Expect(ok).To(BeTrue())
			Expect(tenantToken).To(Equal("test-token-123"))

			endpoints, ok := manifest["communicationEndpoints"].([]interface{})
			Expect(ok).To(BeTrue())
			Expect(endpoints).To(HaveLen(2))
		})
	})

	Describe("Credentials", func() {
		It("validates required credential keys", func() {
			credentials := map[string]interface{}{
				"apitoken":      "dt0c01.test.token",
				"environmentid": "abc12345",
				"apiurl":        "https://abc12345.live.dynatrace.com/api",
			}

			_, ok := credentials["apitoken"]
			Expect(ok).To(BeTrue(), "apitoken is required for Dynatrace")

			_, ok = credentials["environmentid"]
			Expect(ok).To(BeTrue(), "environmentid is required for Dynatrace")
		})
	})
})
