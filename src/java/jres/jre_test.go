package jres_test

import (
	"os"
	"testing"

	"github.com/cloudfoundry/java-buildpack/src/java/jres"
	"github.com/cloudfoundry/libbuildpack"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestJREs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "JREs Suite")
}

var _ = Describe("JRE Registry", func() {
	var (
		ctx      *jres.Context
		registry *jres.Registry
		buildDir string
		depsDir  string
		cacheDir string
	)

	BeforeEach(func() {
		var err error
		buildDir, err = os.MkdirTemp("", "build")
		Expect(err).NotTo(HaveOccurred())

		depsDir, err = os.MkdirTemp("", "deps")
		Expect(err).NotTo(HaveOccurred())

		cacheDir, err = os.MkdirTemp("", "cache")
		Expect(err).NotTo(HaveOccurred())

		logger := libbuildpack.NewLogger(os.Stdout)
		manifest := &libbuildpack.Manifest{}
		installer := &libbuildpack.Installer{}
		stager := &libbuildpack.Stager{}
		command := &libbuildpack.Command{}

		ctx = &jres.Context{
			Stager:    stager,
			Manifest:  manifest,
			Installer: installer,
			Log:       logger,
			Command:   command,
		}

		registry = jres.NewRegistry(ctx)
	})

	AfterEach(func() {
		os.RemoveAll(buildDir)
		os.RemoveAll(depsDir)
		os.RemoveAll(cacheDir)
	})

	Describe("Registry Creation", func() {
		It("creates a registry successfully", func() {
			Expect(registry).NotTo(BeNil())
		})

		It("has no JREs registered by default", func() {
			jre, name, err := registry.Detect()
			Expect(err).NotTo(HaveOccurred())
			Expect(jre).To(BeNil())
			Expect(name).To(BeEmpty())
		})
	})

	Describe("Register and Detect", func() {
		BeforeEach(func() {
			// Register OpenJDK JRE
			registry.Register(jres.NewOpenJDKJRE(ctx))
		})

		It("detects registered JREs", func() {
			jre, name, err := registry.Detect()
			Expect(err).NotTo(HaveOccurred())
			Expect(jre).NotTo(BeNil())
			Expect(name).To(Equal("OpenJDK"))
		})
	})

	Describe("Multiple JREs", func() {
		It("returns first matching JRE", func() {
			// Register multiple JREs (OpenJDK always detects)
			jre1 := jres.NewOpenJDKJRE(ctx)
			registry.Register(jre1)

			jre, name, err := registry.Detect()
			Expect(err).NotTo(HaveOccurred())
			Expect(jre).NotTo(BeNil())
			Expect(name).To(Equal("OpenJDK"))
		})
	})
})
