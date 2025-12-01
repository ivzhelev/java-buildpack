package release_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/cloudfoundry/java-buildpack/src/java/release"
	"github.com/cloudfoundry/libbuildpack"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"
)

func TestRelease(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Release Suite")
}

var _ = Describe("Release", func() {
	var (
		buildDir string
		releaser *release.Releaser
		logger   *libbuildpack.Logger
		stdout   *bytes.Buffer
	)

	BeforeEach(func() {
		var err error

		// Create temp build directory
		buildDir, err = os.MkdirTemp("", "release-build")
		Expect(err).NotTo(HaveOccurred())

		// Create tmp directory for release YAML (simulating finalize phase)
		tmpDir := buildDir + "/tmp"
		err = os.MkdirAll(tmpDir, 0755)
		Expect(err).NotTo(HaveOccurred())

		// Create release YAML file (simulating finalize phase output)
		releaseYamlPath := tmpDir + "/java-buildpack-release-step.yml"
		releaseYaml := `---
default_process_types:
  web: $HOME/.java-buildpack/start.sh
`
		err = os.WriteFile(releaseYamlPath, []byte(releaseYaml), 0644)
		Expect(err).NotTo(HaveOccurred())

		// Create logger with buffer to capture output
		stdout = new(bytes.Buffer)
		logger = libbuildpack.NewLogger(stdout)

		releaser = &release.Releaser{
			BuildDir: buildDir,
			Log:      logger,
		}
	})

	AfterEach(func() {
		os.RemoveAll(buildDir)
	})

	Describe("Run", func() {
		It("outputs valid YAML", func() {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := release.Run(releaser)
			Expect(err).NotTo(HaveOccurred())

			w.Close()
			os.Stdout = oldStdout

			var output bytes.Buffer
			output.ReadFrom(r)

			// Parse the output as YAML
			var releaseInfo map[string]interface{}
			err = yaml.Unmarshal(output.Bytes(), &releaseInfo)
			Expect(err).NotTo(HaveOccurred())
		})

		It("outputs default_process_types", func() {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := release.Run(releaser)
			Expect(err).NotTo(HaveOccurred())

			w.Close()
			os.Stdout = oldStdout

			var output bytes.Buffer
			output.ReadFrom(r)

			outputStr := output.String()
			Expect(outputStr).To(ContainSubstring("default_process_types:"))
		})

		It("specifies web process type", func() {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := release.Run(releaser)
			Expect(err).NotTo(HaveOccurred())

			w.Close()
			os.Stdout = oldStdout

			var output bytes.Buffer
			output.ReadFrom(r)

			outputStr := output.String()
			Expect(outputStr).To(ContainSubstring("web:"))
		})

		It("uses start.sh from .java-buildpack directory", func() {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := release.Run(releaser)
			Expect(err).NotTo(HaveOccurred())

			w.Close()
			os.Stdout = oldStdout

			var output bytes.Buffer
			output.ReadFrom(r)

			outputStr := output.String()
			Expect(outputStr).To(ContainSubstring(".java-buildpack/start.sh"))
		})

		It("starts YAML output with document separator", func() {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := release.Run(releaser)
			Expect(err).NotTo(HaveOccurred())

			w.Close()
			os.Stdout = oldStdout

			var output bytes.Buffer
			output.ReadFrom(r)

			outputStr := output.String()
			Expect(outputStr).To(HavePrefix("---\n"))
		})

		It("references $HOME environment variable", func() {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := release.Run(releaser)
			Expect(err).NotTo(HaveOccurred())

			w.Close()
			os.Stdout = oldStdout

			var output bytes.Buffer
			output.ReadFrom(r)

			outputStr := output.String()
			Expect(outputStr).To(ContainSubstring("$HOME"))
		})
	})

	Describe("YAML Structure", func() {
		It("produces parseable YAML with correct structure", func() {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := release.Run(releaser)
			Expect(err).NotTo(HaveOccurred())

			w.Close()
			os.Stdout = oldStdout

			var output bytes.Buffer
			output.ReadFrom(r)

			// Parse the output as YAML
			var releaseInfo struct {
				DefaultProcessTypes map[string]string `yaml:"default_process_types"`
			}
			err = yaml.Unmarshal(output.Bytes(), &releaseInfo)
			Expect(err).NotTo(HaveOccurred())

			// Verify structure
			Expect(releaseInfo.DefaultProcessTypes).NotTo(BeNil())
			Expect(releaseInfo.DefaultProcessTypes).To(HaveKey("web"))
			Expect(releaseInfo.DefaultProcessTypes["web"]).To(ContainSubstring("start.sh"))
		})
	})
})
