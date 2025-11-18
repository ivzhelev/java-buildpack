package integration_test

import (
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/switchblade"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testSpringBoot(platform switchblade.Platform, fixtures string) func(*testing.T, spec.G, spec.S) {
	return func(t *testing.T, context spec.G, it spec.S) {
		var (
			Expect = NewWithT(t).Expect
			name   string
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

		context("with a Spring Boot application", func() {
			it("successfully deploys and runs", func() {
				deployment, logs, err := platform.Deploy.
					WithEnv(map[string]string{
						"BP_JAVA_VERSION": "11",
					}).
					Execute(name, filepath.Join(fixtures, "container_spring_boot_staged"))
				Expect(err).NotTo(HaveOccurred(), logs.String)

				Expect(logs.String()).To(ContainSubstring("Java Buildpack"))
				Expect(deployment.ExternalURL).NotTo(BeEmpty())
			})
		})

		context("with Spring Auto-reconfiguration", func() {
			it("detects Spring Framework", func() {
				deployment, logs, err := platform.Deploy.
					WithEnv(map[string]string{
						"BP_JAVA_VERSION": "11",
					}).
					Execute(name, filepath.Join(fixtures, "container_spring_boot_staged"))
				Expect(err).NotTo(HaveOccurred(), logs.String)

				// Spring auto-reconfiguration should be detected
				Expect(logs.String()).To(ContainSubstring("Java Buildpack"))
				Expect(deployment.ExternalURL).NotTo(BeEmpty())
			})
		})

		context("with embedded Tomcat", func() {
			it("starts successfully", func() {
				deployment, logs, err := platform.Deploy.
					WithEnv(map[string]string{
						"BP_JAVA_VERSION": "11",
					}).
					Execute(name, filepath.Join(fixtures, "integration_valid"))
				Expect(err).NotTo(HaveOccurred(), logs.String)

				Expect(logs.String()).To(Or(
					ContainSubstring("Tomcat"),
					ContainSubstring("JRE"),
				))
				Expect(deployment.ExternalURL).NotTo(BeEmpty())
			})
		})

		context("with Java CFEnv", func() {
			it("includes java-cfenv when Spring Boot is detected", func() {
				deployment, logs, err := platform.Deploy.
					WithEnv(map[string]string{
						"BP_JAVA_VERSION":       "11",
						"JBP_CONFIG_JAVA_CFENV": "{enabled: true}",
					}).
					Execute(name, filepath.Join(fixtures, "integration_valid"))
				Expect(err).NotTo(HaveOccurred(), logs.String)

				Expect(deployment.ExternalURL).NotTo(BeEmpty())
			})
		})
	}
}
