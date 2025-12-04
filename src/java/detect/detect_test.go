package detect_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/java-buildpack/src/java/detect"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDetect(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Detect Suite")
}

var _ = Describe("Detect", func() {
	var (
		buildDir string
		detector *detect.Detector
	)

	BeforeEach(func() {
		var err error
		buildDir, err = os.MkdirTemp("", "detect-test")
		Expect(err).NotTo(HaveOccurred())

		detector = &detect.Detector{
			BuildDir: buildDir,
			Version:  "1.0.0",
		}
	})

	AfterEach(func() {
		os.RemoveAll(buildDir)
	})

	Context("when detecting servlet applications", func() {
		It("detects WEB-INF directory", func() {
			webInfDir := filepath.Join(buildDir, "WEB-INF")
			Expect(os.MkdirAll(webInfDir, 0755)).To(Succeed())

			err := detect.Run(detector)
			Expect(err).NotTo(HaveOccurred())
		})

		It("detects WAR files", func() {
			warFile := filepath.Join(buildDir, "app.war")
			Expect(os.WriteFile(warFile, []byte("fake war"), 0644)).To(Succeed())

			err := detect.Run(detector)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("when detecting Maven applications", func() {
		It("detects pom.xml", func() {
			pomFile := filepath.Join(buildDir, "pom.xml")
			Expect(os.WriteFile(pomFile, []byte("<project/>"), 0644)).To(Succeed())

			err := detect.Run(detector)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("when detecting Gradle applications", func() {
		It("detects build.gradle", func() {
			gradleFile := filepath.Join(buildDir, "build.gradle")
			Expect(os.WriteFile(gradleFile, []byte("apply plugin: 'java'"), 0644)).To(Succeed())

			err := detect.Run(detector)
			Expect(err).NotTo(HaveOccurred())
		})

		It("detects build.gradle.kts", func() {
			gradleKtsFile := filepath.Join(buildDir, "build.gradle.kts")
			Expect(os.WriteFile(gradleKtsFile, []byte("plugins { java }"), 0644)).To(Succeed())

			err := detect.Run(detector)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("when detecting JAR applications", func() {
		It("detects JAR files", func() {
			jarFile := filepath.Join(buildDir, "app.jar")
			Expect(os.WriteFile(jarFile, []byte("fake jar"), 0644)).To(Succeed())

			err := detect.Run(detector)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("when detecting Spring Boot applications", func() {
		It("detects BOOT-INF directory", func() {
			bootInfDir := filepath.Join(buildDir, "BOOT-INF")
			Expect(os.MkdirAll(bootInfDir, 0755)).To(Succeed())

			err := detect.Run(detector)
			Expect(err).NotTo(HaveOccurred())
		})

		It("detects META-INF/MANIFEST.MF", func() {
			metaInfDir := filepath.Join(buildDir, "META-INF")
			Expect(os.MkdirAll(metaInfDir, 0755)).To(Succeed())
			manifestFile := filepath.Join(metaInfDir, "MANIFEST.MF")
			Expect(os.WriteFile(manifestFile, []byte("Main-Class: com.example.Main"), 0644)).To(Succeed())

			err := detect.Run(detector)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("when detecting class files", func() {
		It("detects .class files", func() {
			classFile := filepath.Join(buildDir, "Main.class")
			Expect(os.WriteFile(classFile, []byte("fake class"), 0644)).To(Succeed())

			err := detect.Run(detector)
			Expect(err).NotTo(HaveOccurred())
		})

		It("detects .class files in subdirectories", func() {
			subDir := filepath.Join(buildDir, "com", "example")
			Expect(os.MkdirAll(subDir, 0755)).To(Succeed())
			classFile := filepath.Join(subDir, "Main.class")
			Expect(os.WriteFile(classFile, []byte("fake class"), 0644)).To(Succeed())

			err := detect.Run(detector)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("when detecting Groovy applications", func() {
		It("detects .groovy files", func() {
			groovyFile := filepath.Join(buildDir, "app.groovy")
			Expect(os.WriteFile(groovyFile, []byte("println 'hello'"), 0644)).To(Succeed())

			err := detect.Run(detector)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("when detecting Play Framework applications", func() {
		It("detects start script at root", func() {
			startScript := filepath.Join(buildDir, "start")
			Expect(os.WriteFile(startScript, []byte("#!/bin/bash"), 0755)).To(Succeed())

			err := detect.Run(detector)
			Expect(err).NotTo(HaveOccurred())
		})

		It("detects start script in application-root", func() {
			appRootDir := filepath.Join(buildDir, "application-root")
			Expect(os.MkdirAll(appRootDir, 0755)).To(Succeed())
			startScript := filepath.Join(appRootDir, "start")
			Expect(os.WriteFile(startScript, []byte("#!/bin/bash"), 0755)).To(Succeed())

			err := detect.Run(detector)
			Expect(err).NotTo(HaveOccurred())
		})

		It("detects start script in staged-app", func() {
			stagedAppDir := filepath.Join(buildDir, "staged-app")
			Expect(os.MkdirAll(stagedAppDir, 0755)).To(Succeed())
			startScript := filepath.Join(stagedAppDir, "start")
			Expect(os.WriteFile(startScript, []byte("#!/bin/bash"), 0755)).To(Succeed())

			err := detect.Run(detector)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("when detecting Ratpack applications", func() {
		It("detects ratpack-core JAR", func() {
			libDir := filepath.Join(buildDir, "application-root", "lib")
			Expect(os.MkdirAll(libDir, 0755)).To(Succeed())
			ratpackJar := filepath.Join(libDir, "ratpack-core-1.5.0.jar")
			Expect(os.WriteFile(ratpackJar, []byte("fake jar"), 0644)).To(Succeed())

			err := detect.Run(detector)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("when detecting generic Java applications", func() {
		It("detects application-root/lib with JARs", func() {
			libDir := filepath.Join(buildDir, "application-root", "lib")
			Expect(os.MkdirAll(libDir, 0755)).To(Succeed())
			jarFile := filepath.Join(libDir, "app.jar")
			Expect(os.WriteFile(jarFile, []byte("fake jar"), 0644)).To(Succeed())

			err := detect.Run(detector)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("when detecting dist-zip applications", func() {
		It("detects bin/ and lib/ directories at root", func() {
			binDir := filepath.Join(buildDir, "bin")
			libDir := filepath.Join(buildDir, "lib")
			Expect(os.MkdirAll(binDir, 0755)).To(Succeed())
			Expect(os.MkdirAll(libDir, 0755)).To(Succeed())

			// Create a non-.bat script in bin/
			startScript := filepath.Join(binDir, "start")
			Expect(os.WriteFile(startScript, []byte("#!/bin/bash"), 0755)).To(Succeed())

			err := detect.Run(detector)
			Expect(err).NotTo(HaveOccurred())
		})

		It("detects bin/ and lib/ directories in application-root", func() {
			appRoot := filepath.Join(buildDir, "application-root")
			binDir := filepath.Join(appRoot, "bin")
			libDir := filepath.Join(appRoot, "lib")
			Expect(os.MkdirAll(binDir, 0755)).To(Succeed())
			Expect(os.MkdirAll(libDir, 0755)).To(Succeed())

			// Create a non-.bat script in bin/
			startScript := filepath.Join(binDir, "start")
			Expect(os.WriteFile(startScript, []byte("#!/bin/bash"), 0755)).To(Succeed())

			err := detect.Run(detector)
			Expect(err).NotTo(HaveOccurred())
		})

		It("ignores .bat files in bin/ directory", func() {
			binDir := filepath.Join(buildDir, "bin")
			libDir := filepath.Join(buildDir, "lib")
			Expect(os.MkdirAll(binDir, 0755)).To(Succeed())
			Expect(os.MkdirAll(libDir, 0755)).To(Succeed())

			// Create only .bat script in bin/
			batScript := filepath.Join(binDir, "start.bat")
			Expect(os.WriteFile(batScript, []byte("@echo off"), 0644)).To(Succeed())

			err := detect.Run(detector)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when detecting Procfile", func() {
		It("detects Procfile with content", func() {
			procfile := filepath.Join(buildDir, "Procfile")
			Expect(os.WriteFile(procfile, []byte("web: java -jar app.jar"), 0644)).To(Succeed())

			err := detect.Run(detector)
			Expect(err).NotTo(HaveOccurred())
		})

		It("fails when Procfile is empty", func() {
			procfile := filepath.Join(buildDir, "Procfile")
			Expect(os.WriteFile(procfile, []byte(""), 0644)).To(Succeed())

			err := detect.Run(detector)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when no Java application is detected", func() {
		It("returns an error", func() {
			err := detect.Run(detector)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no Java app detected"))
		})
	})
})
