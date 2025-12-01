package release

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack"
)

type Releaser struct {
	BuildDir string
	Manifest *libbuildpack.Manifest
	Log      *libbuildpack.Logger
}

// Run generates the release information
// This follows the reference buildpack pattern (Ruby, Go, Node.js, Python):
// 1. Read the YAML file written by the finalize phase
// 2. Output it to stdout for Cloud Foundry to parse
// The YAML file contains the direct container command (e.g., bin/application for DistZip)
func Run(r *Releaser) error {
	releaseYamlPath := filepath.Join(r.BuildDir, "tmp", "java-buildpack-release-step.yml")

	// Read the YAML file written by finalize phase
	yamlContent, err := os.ReadFile(releaseYamlPath)
	if err != nil {
		r.Log.Error("Failed to read release YAML file: %s", err.Error())
		return fmt.Errorf("reading release YAML: %w", err)
	}

	// Output the YAML content to stdout
	// Cloud Foundry will parse this to determine the web command
	fmt.Print(string(yamlContent))

	return nil
}
