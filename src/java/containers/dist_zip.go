package containers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DistZipContainer handles distribution ZIP applications
// (applications with bin/ and lib/ structure, typically from Gradle's distZip)
type DistZipContainer struct {
	context     *Context
	startScript string
}

// NewDistZipContainer creates a new Dist ZIP container
func NewDistZipContainer(ctx *Context) *DistZipContainer {
	return &DistZipContainer{
		context: ctx,
	}
}

// Detect checks if this is a Dist ZIP application
func (d *DistZipContainer) Detect() (string, error) {
	buildDir := d.context.Stager.BuildDir()

	// Check for bin/ and lib/ directories (typical distZip structure)
	binDir := filepath.Join(buildDir, "bin")
	libDir := filepath.Join(buildDir, "lib")

	binStat, binErr := os.Stat(binDir)
	libStat, libErr := os.Stat(libDir)

	if binErr == nil && libErr == nil && binStat.IsDir() && libStat.IsDir() {
		// Check for startup scripts in bin/
		entries, err := os.ReadDir(binDir)
		if err == nil && len(entries) > 0 {
			// Find a non-.bat script (Unix startup script)
			for _, entry := range entries {
				if !entry.IsDir() && filepath.Ext(entry.Name()) != ".bat" {
					d.startScript = entry.Name()
					d.context.Log.Debug("Detected Dist ZIP application with start script: %s", d.startScript)
					return "Dist ZIP", nil
				}
			}
		}
	}

	return "", nil
}

// Supply installs Dist ZIP dependencies
func (d *DistZipContainer) Supply() error {
	d.context.Log.BeginStep("Supplying Dist ZIP")

	// For Dist ZIP apps, the structure is already provided
	// We may need to:
	// 1. Ensure scripts are executable
	// 2. Install support utilities

	// Make bin scripts executable
	if err := d.makeScriptsExecutable(); err != nil {
		d.context.Log.Warning("Could not make scripts executable: %s", err.Error())
	}

	// Install JVMKill agent
	if err := d.installJVMKillAgent(); err != nil {
		d.context.Log.Warning("Could not install JVMKill agent: %s", err.Error())
	}

	return nil
}

// makeScriptsExecutable ensures all scripts in bin/ are executable
func (d *DistZipContainer) makeScriptsExecutable() error {
	binDir := filepath.Join(d.context.Stager.BuildDir(), "bin")

	entries, err := os.ReadDir(binDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) != ".bat" {
			scriptPath := filepath.Join(binDir, entry.Name())
			if err := os.Chmod(scriptPath, 0755); err != nil {
				d.context.Log.Warning("Could not make %s executable: %s", entry.Name(), err.Error())
			}
		}
	}

	return nil
}

// installJVMKillAgent installs the JVMKill agent
func (d *DistZipContainer) installJVMKillAgent() error {
	dep, err := d.context.Manifest.DefaultVersion("jvmkill")
	if err != nil {
		return err
	}

	jvmkillPath := filepath.Join(d.context.Stager.DepDir(), "jvmkill")
	if err := d.context.Installer.InstallDependency(dep, jvmkillPath); err != nil {
		return fmt.Errorf("failed to install JVMKill: %w", err)
	}

	d.context.Log.Info("Installed JVMKill agent version %s", dep.Version)
	return nil
}

// Finalize performs final Dist ZIP configuration
func (d *DistZipContainer) Finalize() error {
	d.context.Log.BeginStep("Finalizing Dist ZIP")

	// Configure JAVA_OPTS to be picked up by startup scripts
	javaOpts := []string{
		"-Djava.io.tmpdir=$TMPDIR",
		"-XX:+ExitOnOutOfMemoryError",
	}

	// Add JVMKill agent if available
	jvmkillSO := filepath.Join(d.context.Stager.DepDir(), "jvmkill", "jvmkill.so")
	if _, err := os.Stat(jvmkillSO); err == nil {
		javaOpts = append(javaOpts, fmt.Sprintf("-agentpath:%s", jvmkillSO))
	}

	// Most distZip scripts respect JAVA_OPTS environment variable
	// Write JAVA_OPTS for the startup script to use
	if err := d.context.Stager.WriteEnvFile("JAVA_OPTS",
		strings.Join(javaOpts, " ")); err != nil {
		return fmt.Errorf("failed to write JAVA_OPTS: %w", err)
	}

	return nil
}

// Release returns the Dist ZIP startup command
func (d *DistZipContainer) Release() (string, error) {
	// Use the detected start script
	if d.startScript == "" {
		// Try to detect again
		if _, err := d.Detect(); err != nil || d.startScript == "" {
			return "", fmt.Errorf("no start script found in bin/ directory")
		}
	}

	cmd := filepath.Join("bin", d.startScript)
	return cmd, nil
}
