package containers_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/java-buildpack/src/java/containers"
	"github.com/cloudfoundry/libbuildpack"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestContainers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Containers Suite")
}

var _ = Describe("Container Registry", func() {
	var (
		ctx      *containers.Context
		registry *containers.Registry
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

		// Create deps directory structure
		err = os.MkdirAll(filepath.Join(depsDir, "0"), 0755)
		Expect(err).NotTo(HaveOccurred())

		logger := libbuildpack.NewLogger(os.Stdout)
		manifest := &libbuildpack.Manifest{}
		installer := &libbuildpack.Installer{}
		stager := libbuildpack.NewStager([]string{buildDir, cacheDir, depsDir, "0"}, logger, manifest)
		command := &libbuildpack.Command{}

		ctx = &containers.Context{
			Stager:    stager,
			Manifest:  manifest,
			Installer: installer,
			Log:       logger,
			Command:   command,
		}

		registry = containers.NewRegistry(ctx)
	})

	AfterEach(func() {
		os.RemoveAll(buildDir)
		os.RemoveAll(depsDir)
		os.RemoveAll(cacheDir)
	})

	Describe("Spring Boot Container", func() {
		Context("with BOOT-INF directory", func() {
			BeforeEach(func() {
				os.MkdirAll(filepath.Join(buildDir, "BOOT-INF"), 0755)
			})

			It("detects as Spring Boot", func() {
				container := containers.NewSpringBootContainer(ctx)
				name, err := container.Detect()
				Expect(err).NotTo(HaveOccurred())
				Expect(name).To(Equal("Spring Boot"))
			})
		})

		Context("without Spring Boot indicators", func() {
			It("does not detect as Spring Boot", func() {
				container := containers.NewSpringBootContainer(ctx)
				name, err := container.Detect()
				Expect(err).NotTo(HaveOccurred())
				Expect(name).To(BeEmpty())
			})
		})
	})

	Describe("Tomcat Container", func() {
		Context("with WEB-INF directory", func() {
			BeforeEach(func() {
				os.MkdirAll(filepath.Join(buildDir, "WEB-INF"), 0755)
			})

			It("detects as Tomcat", func() {
				container := containers.NewTomcatContainer(ctx)
				name, err := container.Detect()
				Expect(err).NotTo(HaveOccurred())
				Expect(name).To(Equal("Tomcat"))
			})
		})

		Context("with WAR file", func() {
			BeforeEach(func() {
				os.WriteFile(filepath.Join(buildDir, "app.war"), []byte{}, 0644)
			})

			It("detects as Tomcat", func() {
				container := containers.NewTomcatContainer(ctx)
				name, err := container.Detect()
				Expect(err).NotTo(HaveOccurred())
				Expect(name).To(Equal("Tomcat"))
			})
		})
	})

	Describe("Groovy Container", func() {
		Context("with .groovy files", func() {
			BeforeEach(func() {
				os.WriteFile(filepath.Join(buildDir, "app.groovy"), []byte("println 'hello'"), 0644)
			})

			It("detects as Groovy", func() {
				container := containers.NewGroovyContainer(ctx)
				name, err := container.Detect()
				Expect(err).NotTo(HaveOccurred())
				Expect(name).To(Equal("Groovy"))
			})
		})
	})

	Describe("Dist ZIP Container", func() {
		Context("with bin/ and lib/ directories and startup script", func() {
			BeforeEach(func() {
				os.MkdirAll(filepath.Join(buildDir, "bin"), 0755)
				os.MkdirAll(filepath.Join(buildDir, "lib"), 0755)
				os.WriteFile(filepath.Join(buildDir, "bin", "start"), []byte("#!/bin/sh"), 0755)
			})

			It("detects as Dist ZIP", func() {
				container := containers.NewDistZipContainer(ctx)
				name, err := container.Detect()
				Expect(err).NotTo(HaveOccurred())
				Expect(name).To(Equal("Dist ZIP"))
			})
		})
	})

	Describe("Java Main Container", func() {
		Context("with JAR file", func() {
			BeforeEach(func() {
				os.WriteFile(filepath.Join(buildDir, "app.jar"), []byte{}, 0644)
			})

			It("detects as Java Main", func() {
				container := containers.NewJavaMainContainer(ctx)
				name, err := container.Detect()
				Expect(err).NotTo(HaveOccurred())
				Expect(name).To(Equal("Java Main"))
			})
		})

		Context("with .class files", func() {
			BeforeEach(func() {
				os.WriteFile(filepath.Join(buildDir, "Main.class"), []byte{}, 0644)
			})

			It("detects as Java Main", func() {
				container := containers.NewJavaMainContainer(ctx)
				name, err := container.Detect()
				Expect(err).NotTo(HaveOccurred())
				Expect(name).To(Equal("Java Main"))
			})
		})
	})

	Describe("Registry", func() {
		BeforeEach(func() {
			registry.Register(containers.NewSpringBootContainer(ctx))
			registry.Register(containers.NewTomcatContainer(ctx))
			registry.Register(containers.NewGroovyContainer(ctx))
			registry.Register(containers.NewDistZipContainer(ctx))
			registry.Register(containers.NewJavaMainContainer(ctx))
		})

		Context("with Spring Boot app", func() {
			BeforeEach(func() {
				os.MkdirAll(filepath.Join(buildDir, "BOOT-INF"), 0755)
			})

			It("detects Spring Boot container", func() {
				container, name, err := registry.Detect()
				Expect(err).NotTo(HaveOccurred())
				Expect(container).NotTo(BeNil())
				Expect(name).To(Equal("Spring Boot"))
			})
		})

		Context("with no detectable app", func() {
			It("returns nil container", func() {
				container, name, err := registry.Detect()
				Expect(err).NotTo(HaveOccurred())
				Expect(container).To(BeNil())
				Expect(name).To(BeEmpty())
			})
		})
	})
})
