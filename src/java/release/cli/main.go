package main

import (
	"fmt"
	"os"
	"time"

	"github.com/cloudfoundry/java-buildpack/src/java/release"
	"github.com/cloudfoundry/libbuildpack"
)

func main() {
	// Release phase only takes BUILD_DIR as argument
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: release <build-dir>")
		os.Exit(1)
	}

	buildDir := os.Args[1]

	logger := libbuildpack.NewLogger(os.Stdout)

	buildpackDir, err := libbuildpack.GetBuildpackDir()
	if err != nil {
		logger.Error("Unable to determine buildpack directory: %s", err.Error())
		os.Exit(9)
	}

	manifest, err := libbuildpack.NewManifest(buildpackDir, logger, time.Now())
	if err != nil {
		logger.Error("Unable to load buildpack manifest: %s", err.Error())
		os.Exit(10)
	}

	r := release.Releaser{
		BuildDir: buildDir,
		Manifest: manifest,
		Log:      logger,
	}

	if err := release.Run(&r); err != nil {
		logger.Error("Failed to generate release: %s", err.Error())
		os.Exit(14)
	}
}
