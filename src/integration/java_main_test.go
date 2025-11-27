package integration_test

import (
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/switchblade"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testJavaMain(platform switchblade.Platform, fixtures string) func(*testing.T, spec.G, spec.S) {
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
			if name != "" {
				Expect(platform.Delete.Execute(name)).To(Succeed())
			}
		})

		context("with a Java Main application", func() {
			it("successfully deploys with Main-Class manifest entry", func() {
				deployment, logs, err := platform.Deploy.
					WithEnv(map[string]string{
						"BP_JAVA_VERSION": "11",
					}).
					Execute(name, filepath.Join(fixtures, "container_main"))
				Expect(err).NotTo(HaveOccurred(), logs.String)

				// Should detect Main-Class from MANIFEST.MF
				Expect(logs.String()).To(ContainSubstring("Java Buildpack"))
				Eventually(deployment.ExternalURL).Should(Not(BeEmpty()))
			})
		})

		context("with explicit main class", func() {
			it("uses the specified main class", func() {
				deployment, logs, err := platform.Deploy.
					WithEnv(map[string]string{
						"BP_JAVA_VERSION":      "11",
						"JBP_CONFIG_JAVA_MAIN": `{java_main_class: "io.pivotal.SimpleJava"}`,
					}).
					Execute(name, filepath.Join(fixtures, "container_main"))
				Expect(err).NotTo(HaveOccurred(), logs.String)

				Expect(logs.String()).To(ContainSubstring("Java Buildpack"))
				Eventually(deployment.ExternalURL).Should(Not(BeEmpty()))
			})
		})

		context("with custom arguments", func() {
			it("passes arguments to the main class", func() {
				deployment, logs, err := platform.Deploy.
					WithEnv(map[string]string{
						"BP_JAVA_VERSION":      "11",
						"JBP_CONFIG_JAVA_MAIN": `{arguments: "--server.port=$PORT"}`,
					}).
					Execute(name, filepath.Join(fixtures, "integration_valid"))
				Expect(err).NotTo(HaveOccurred(), logs.String)

				Eventually(deployment.ExternalURL).Should(Not(BeEmpty()))
			})
		})

		context("with JAVA_OPTS", func() {
			it("applies custom Java options", func() {
				deployment, logs, err := platform.Deploy.
					WithEnv(map[string]string{
						"BP_JAVA_VERSION": "11",
						"JAVA_OPTS":       "-Xmx512m -XX:+UseG1GC",
					}).
					Execute(name, filepath.Join(fixtures, "container_main"))
				Expect(err).NotTo(HaveOccurred(), logs.String)

				Expect(logs.String()).To(ContainSubstring("Java Buildpack"))
				Eventually(deployment.ExternalURL).Should(Not(BeEmpty()))
			})
		})
	}
}
