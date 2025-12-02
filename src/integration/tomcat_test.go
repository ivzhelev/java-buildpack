package integration_test

import (
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/switchblade"
	"github.com/cloudfoundry/switchblade/matchers"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testTomcat(platform switchblade.Platform, fixtures string) func(*testing.T, spec.G, spec.S) {
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

		context("with a simple servlet app", func() {
			it("successfully deploys and runs with Java 11 (Jakarta EE)", func() {
				deployment, logs, err := platform.Deploy.
					WithEnv(map[string]string{
						"BP_JAVA_VERSION": "11",
					}).
					Execute(name, filepath.Join(fixtures, "container_tomcat_jakarta"))

				Expect(err).NotTo(HaveOccurred(), logs.String)

				Eventually(deployment).Should(matchers.Serve(ContainSubstring("OK")))
			})
		})

		context("with JRE selection", func() {
			it("deploys with Java 8 (Tomcat 9 + javax.servlet)", func() {
				deployment, logs, err := platform.Deploy.
					WithEnv(map[string]string{
						"BP_JAVA_VERSION": "8",
					}).
					Execute(name, filepath.Join(fixtures, "container_tomcat_javax"))
				Expect(err).NotTo(HaveOccurred(), logs.String)

				Expect(logs.String()).To(ContainSubstring("OpenJDK"))
				Expect(logs.String()).To(ContainSubstring("Tomcat 9"))
				Eventually(deployment).Should(matchers.Serve(ContainSubstring("OK")))
			})

			it("deploys with Java 11 (Tomcat 10 + jakarta.servlet)", func() {
				deployment, logs, err := platform.Deploy.
					WithEnv(map[string]string{
						"BP_JAVA_VERSION": "11",
					}).
					Execute(name, filepath.Join(fixtures, "container_tomcat_jakarta"))
				Expect(err).NotTo(HaveOccurred(), logs.String)

				Expect(logs.String()).To(ContainSubstring("OpenJDK"))
				Expect(logs.String()).To(ContainSubstring("Tomcat 10"))
				Eventually(deployment).Should(matchers.Serve(ContainSubstring("OK")))
			})

			it("deploys with Java 17 (Tomcat 10 + jakarta.servlet)", func() {
				deployment, logs, err := platform.Deploy.
					WithEnv(map[string]string{
						"BP_JAVA_VERSION": "17",
					}).
					Execute(name, filepath.Join(fixtures, "container_tomcat_jakarta"))
				Expect(err).NotTo(HaveOccurred(), logs.String)

				Expect(logs.String()).To(ContainSubstring("OpenJDK"))
				Expect(logs.String()).To(ContainSubstring("Tomcat 10"))
				Eventually(deployment).Should(matchers.Serve(ContainSubstring("OK")))
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
					Execute(name, filepath.Join(fixtures, "container_tomcat_jakarta"))

				Expect(err).NotTo(HaveOccurred(), logs.String)

				Expect(logs.String()).To(ContainSubstring("memory"))
				Eventually(deployment).Should(matchers.Serve(ContainSubstring("OK")))
			})
		})
	}
}
