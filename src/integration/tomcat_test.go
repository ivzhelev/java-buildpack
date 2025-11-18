package integration_test

import (
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/switchblade"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testTomcat(platform switchblade.Platform, fixtures string) func(*testing.T, spec.G, spec.S) {
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

		context("with a simple servlet app", func() {
			it("successfully deploys and runs", func() {
				deployment, logs, err := platform.Deploy.
					WithEnv(map[string]string{
						"BP_JAVA_VERSION": "8",
					}).
					Execute(name, filepath.Join(fixtures, "integration_valid"))
				Expect(err).NotTo(HaveOccurred(), logs.String)

				Expect(deployment.ExternalURL).NotTo(BeEmpty())
			})
		})

		context("with JRE selection", func() {
			it("deploys with Java 8", func() {
				deployment, logs, err := platform.Deploy.
					WithEnv(map[string]string{
						"BP_JAVA_VERSION": "8",
					}).
					Execute(name, filepath.Join(fixtures, "integration_valid"))
				Expect(err).NotTo(HaveOccurred(), logs.String)

				Expect(logs.String()).To(ContainSubstring("Open Jdk JRE"))
				Expect(deployment.ExternalURL).NotTo(BeEmpty())
			})

			it("deploys with Java 11", func() {
				deployment, logs, err := platform.Deploy.
					WithEnv(map[string]string{
						"BP_JAVA_VERSION": "11",
					}).
					Execute(name, filepath.Join(fixtures, "integration_valid"))
				Expect(err).NotTo(HaveOccurred(), logs.String)

				Expect(logs.String()).To(ContainSubstring("Open Jdk JRE"))
				Expect(deployment.ExternalURL).NotTo(BeEmpty())
			})

			it("deploys with Java 17", func() {
				deployment, logs, err := platform.Deploy.
					WithEnv(map[string]string{
						"BP_JAVA_VERSION": "17",
					}).
					Execute(name, filepath.Join(fixtures, "integration_valid"))
				Expect(err).NotTo(HaveOccurred(), logs.String)

				Expect(logs.String()).To(ContainSubstring("Open Jdk JRE"))
				Expect(deployment.ExternalURL).NotTo(BeEmpty())
			})
		})

		context("with memory limits", func() {
			it("respects memory calculator settings", func() {
				deployment, logs, err := platform.Deploy.
					WithEnv(map[string]string{
						"BP_JAVA_VERSION":         "11",
						"JAVA_OPTS":               "-Xmx256m",
						"JBP_CONFIG_OPEN_JDK_JRE": "{jre: {version: 11.+}}",
					}).
					Execute(name, filepath.Join(fixtures, "integration_valid"))
				Expect(err).NotTo(HaveOccurred(), logs.String)

				Expect(logs.String()).To(ContainSubstring("memory"))
				Expect(deployment.ExternalURL).NotTo(BeEmpty())
			})
		})
	}
}
