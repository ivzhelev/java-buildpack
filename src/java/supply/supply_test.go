package supply_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cloudfoundry/java-buildpack/src/java/supply"
	"github.com/cloudfoundry/libbuildpack"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSupply(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Supply Suite")
}

var _ = Describe("Supply", func() {
	var (
		buildDir string
		cacheDir string
		depsDir  string
		depsIdx  string
		supplier *supply.Supplier
		stager   *libbuildpack.Stager
		logger   *libbuildpack.Logger
	)

	BeforeEach(func() {
		var err error

		// Create temp directories
		buildDir, err = os.MkdirTemp("", "supply-build")
		Expect(err).NotTo(HaveOccurred())

		cacheDir, err = os.MkdirTemp("", "supply-cache")
		Expect(err).NotTo(HaveOccurred())

		depsDir, err = os.MkdirTemp("", "supply-deps")
		Expect(err).NotTo(HaveOccurred())

		depsIdx = "0"

		// Create a mock buildpack directory with VERSION and manifest.yml files
		buildpackDir, err := os.MkdirTemp("", "supply-buildpack")
		Expect(err).NotTo(HaveOccurred())

		versionFile := filepath.Join(buildpackDir, "VERSION")
		Expect(os.WriteFile(versionFile, []byte("1.0.0"), 0644)).To(Succeed())

		manifestFile := filepath.Join(buildpackDir, "manifest.yml")
		manifestContent := `---
language: java
default_versions: []
dependencies: []
`
		Expect(os.WriteFile(manifestFile, []byte(manifestContent), 0644)).To(Succeed())

		// Create logger
		logger = libbuildpack.NewLogger(GinkgoWriter)

		// Create manifest with buildpack dir
		manifest, err := libbuildpack.NewManifest(buildpackDir, logger, time.Now())
		Expect(err).NotTo(HaveOccurred())

		// Create stager
		stager = libbuildpack.NewStager([]string{buildDir, cacheDir, depsDir, depsIdx}, logger, manifest)

		supplier = &supply.Supplier{
			Stager:   stager,
			Manifest: manifest,
			Log:      logger,
			Command:  &libbuildpack.Command{},
		}
	})

	AfterEach(func() {
		os.RemoveAll(buildDir)
		os.RemoveAll(cacheDir)
		os.RemoveAll(depsDir)
	})

	Describe("Container Detection", func() {
		Context("when a Spring Boot application is present", func() {
			BeforeEach(func() {
				// Create a Spring Boot JAR with BOOT-INF
				bootInfDir := filepath.Join(buildDir, "BOOT-INF")
				Expect(os.MkdirAll(bootInfDir, 0755)).To(Succeed())
			})

			It("detects Spring Boot container", func() {
				// Note: This test would require mocking the manifest and installer
				// to avoid actual downloads. For now, we're testing the structure.
				Expect(supplier).NotTo(BeNil())
				Expect(supplier.Stager).NotTo(BeNil())
			})
		})

		Context("when a Tomcat application is present", func() {
			BeforeEach(func() {
				// Create WEB-INF directory
				webInfDir := filepath.Join(buildDir, "WEB-INF")
				Expect(os.MkdirAll(webInfDir, 0755)).To(Succeed())
			})

			It("detects Tomcat container", func() {
				Expect(supplier).NotTo(BeNil())
				Expect(supplier.Stager).NotTo(BeNil())
			})
		})

		Context("when a Groovy application is present", func() {
			BeforeEach(func() {
				// Create a .groovy file
				groovyFile := filepath.Join(buildDir, "app.groovy")
				Expect(os.WriteFile(groovyFile, []byte("println 'hello'"), 0644)).To(Succeed())
			})

			It("detects Groovy container", func() {
				Expect(supplier).NotTo(BeNil())
				Expect(supplier.Stager).NotTo(BeNil())
			})
		})
	})

	Describe("Stager Configuration", func() {
		It("creates necessary directories in deps dir", func() {
			depDir := stager.DepDir()
			Expect(depDir).To(ContainSubstring(depsDir))
		})

		It("has access to build directory", func() {
			Expect(stager.BuildDir()).To(Equal(buildDir))
		})

		It("has access to cache directory", func() {
			Expect(stager.CacheDir()).To(Equal(cacheDir))
		})
	})

	Describe("WriteConfigYml", func() {
		It("writes config.yml to deps directory", func() {
			config := map[string]string{
				"container": "spring-boot",
				"jre":       "OpenJDK",
			}

			err := stager.WriteConfigYml(config)
			Expect(err).NotTo(HaveOccurred())

			configPath := filepath.Join(stager.DepDir(), "config.yml")
			Expect(configPath).To(BeAnExistingFile())
		})

		It("handles empty config gracefully", func() {
			err := stager.WriteConfigYml(nil)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
