package release

import (
	"fmt"

	"github.com/cloudfoundry/libbuildpack"
)

type Releaser struct {
	BuildDir string
	Manifest *libbuildpack.Manifest
	Log      *libbuildpack.Logger
}

// Run generates the release information
// The release phase is simple - it just outputs the default process type (web command)
// The actual startup command will be determined at runtime by the finalized container
func Run(r *Releaser) error {
	// Output default process types in YAML format
	// This must be valid YAML parseable by Cloud Foundry
	fmt.Println("---")
	fmt.Println("default_process_types:")
	fmt.Println("  web: $HOME/.java-buildpack/start.sh")
	return nil
}
