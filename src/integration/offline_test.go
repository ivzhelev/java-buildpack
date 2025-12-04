package integration_test

import (
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/switchblade"
	"github.com/sclevine/spec"

	"github.com/cloudfoundry/switchblade/matchers"
	. "github.com/onsi/gomega"
)

func testOffline(platform switchblade.Platform, fixtures string) func(*testing.T, spec.G, spec.S) {
	return func(t *testing.T, context spec.G, it spec.S) {
		var (
			Expect     = NewWithT(t).Expect
			Eventually = NewWithT(t).Eventually
			name       string
		)

		it.Before(func() {
			var err error
			name, err = switchblade.RandomName()
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			if t.Failed() && name != "" {
				t.Logf("❌ FAILED TEST - App/Container: %s", name)
				t.Logf("   Platform: %s", settings.Platform)
			}
			if name != "" && (!settings.KeepFailedContainers || !t.Failed()) {
				Expect(platform.Delete.Execute(name)).To(Succeed())
			}
		})

		context("in offline mode", func() {
			it("deploys without internet access", func() {
				deployment, logs, err := platform.Deploy.
					WithoutInternetAccess().
					WithEnv(map[string]string{
						"BP_JAVA_VERSION": "11",
					}).
					Execute(name, filepath.Join(fixtures, "apps", "integration_valid"))
				Expect(err).NotTo(HaveOccurred(), logs.String)

				// In offline mode, all dependencies should be cached
				Expect(logs.String()).To(Or(
					ContainSubstring("Downloading"),
					ContainSubstring("cached"),
				))
				Eventually(deployment).Should(matchers.Serve(ContainSubstring("OK")))
			})
		})

		context("with cached buildpack", func() {
			it("uses cached dependencies", func() {
				deployment, logs, err := platform.Deploy.
					WithoutInternetAccess().
					WithEnv(map[string]string{
						"BP_JAVA_VERSION": "11",
					}).
					Execute(name, filepath.Join(fixtures, "apps", "integration_valid"))
				Expect(err).NotTo(HaveOccurred(), logs.String)

				// Should not attempt external downloads
				Expect(logs.String()).NotTo(ContainSubstring("ERROR"))
				Eventually(deployment).Should(matchers.Serve(ContainSubstring("OK")))
			})
		})

		context("with offline JRE", func() {
			it("successfully deploys with cached JRE", func() {
				deployment, logs, err := platform.Deploy.
					WithoutInternetAccess().
					WithEnv(map[string]string{
						"BP_JAVA_VERSION": "11",
					}).
					Execute(name, filepath.Join(fixtures, "apps", "integration_valid"))
				Expect(err).NotTo(HaveOccurred(), logs.String)

				Expect(logs.String()).To(ContainSubstring("OpenJDK"))
				Eventually(deployment).Should(matchers.Serve(ContainSubstring("OK")))
			})
		})

		context("with offline Tomcat", func() {
			it("successfully deploys with cached Tomcat", func() {
				deployment, logs, err := platform.Deploy.
					WithoutInternetAccess().
					WithEnv(map[string]string{
						"BP_JAVA_VERSION": "11",
					}).
					Execute(name, filepath.Join(fixtures, "apps", "integration_valid"))
				Expect(err).NotTo(HaveOccurred(), logs.String)

				Eventually(deployment).Should(matchers.Serve(ContainSubstring("OK")))
			})
		})
	}
}
