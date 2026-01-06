package frameworks_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TakipiAgent", func() {
	It("should require secret_key credential", func() {
		credentials := map[string]interface{}{
			"secret_key": "test-secret-key-xyz",
		}

		key, ok := credentials["secret_key"].(string)
		Expect(ok).To(BeTrue())
		Expect(key).NotTo(BeEmpty())
	})
})
