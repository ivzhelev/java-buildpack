package integration_test

import (
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/switchblade"
	"github.com/sclevine/spec"

	"github.com/cloudfoundry/switchblade/matchers"
	. "github.com/onsi/gomega"
)

func testGroovy(platform switchblade.Platform, fixtures string) func(*testing.T, spec.G, spec.S) {
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

		context("with a simple Groovy application", func() {
			it("successfully deploys a non-POGO Groovy script", func() {
				deployment, logs, err := platform.Deploy.
					WithEnv(map[string]string{
						"BP_JAVA_VERSION": "11",
					}).
					Execute(name, filepath.Join(fixtures, "containers", "groovy_non_pogo"))
				Expect(err).NotTo(HaveOccurred(), logs.String)

				Expect(logs.String()).To(ContainSubstring("Java Buildpack"))
				Eventually(deployment).Should(matchers.Serve(ContainSubstring("Hello World")))
			})

			it("successfully deploys a Groovy script with main method", func() {
				deployment, logs, err := platform.Deploy.
					WithEnv(map[string]string{
						"BP_JAVA_VERSION": "11",
					}).
					Execute(name, filepath.Join(fixtures, "containers", "groovy_main_method"))
				Expect(err).NotTo(HaveOccurred(), logs.String)

				Expect(logs.String()).To(ContainSubstring("Java Buildpack"))
				Eventually(deployment).Should(matchers.Serve(ContainSubstring("Class Path")))
			})

			it("successfully deploys a Groovy script with shebang", func() {
				deployment, logs, err := platform.Deploy.
					WithEnv(map[string]string{
						"BP_JAVA_VERSION": "11",
					}).
					Execute(name, filepath.Join(fixtures, "containers", "groovy_shebang"))
				Expect(err).NotTo(HaveOccurred(), logs.String)

				Expect(logs.String()).To(ContainSubstring("Java Buildpack"))
				Eventually(deployment).Should(matchers.Serve(Not(BeEmpty())))
			})
		})

		context("with Groovy and JAR files", func() {
			it("successfully deploys when JARs are present", func() {
				deployment, logs, err := platform.Deploy.
					WithEnv(map[string]string{
						"BP_JAVA_VERSION": "11",
					}).
					Execute(name, filepath.Join(fixtures, "containers", "groovy_with_jars"))
				Expect(err).NotTo(HaveOccurred(), logs.String)

				Expect(logs.String()).To(ContainSubstring("Java Buildpack"))
				Eventually(deployment).Should(matchers.Serve(Not(BeEmpty())))
			})
		})

		context("with edge cases", func() {
			it("successfully deploys Groovy script with shebang containing class", func() {
				deployment, logs, err := platform.Deploy.
					WithEnv(map[string]string{
						"BP_JAVA_VERSION": "11",
					}).
					Execute(name, filepath.Join(fixtures, "containers", "groovy_shebang_containing_class"))
				Expect(err).NotTo(HaveOccurred(), logs.String)

				Expect(logs.String()).To(ContainSubstring("Java Buildpack"))
				Eventually(deployment).Should(matchers.Serve(Not(BeEmpty())))
			})
		})
	}
}
