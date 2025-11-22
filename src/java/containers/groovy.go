package containers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GroovyContainer handles Groovy script applications
type GroovyContainer struct {
	context       *Context
	groovyScripts []string
}

// NewGroovyContainer creates a new Groovy container
func NewGroovyContainer(ctx *Context) *GroovyContainer {
	return &GroovyContainer{
		context: ctx,
	}
}

// Detect checks if this is a Groovy application
func (g *GroovyContainer) Detect() (string, error) {
	buildDir := g.context.Stager.BuildDir()

	// Look for .groovy files
	groovyFiles, err := filepath.Glob(filepath.Join(buildDir, "*.groovy"))
	if err != nil {
		return "", err
	}

	if len(groovyFiles) > 0 {
		g.groovyScripts = groovyFiles
		g.context.Log.Debug("Detected Groovy application with %d script(s)", len(groovyFiles))
		return "Groovy", nil
	}

	return "", nil
}

// Supply installs Groovy and dependencies
func (g *GroovyContainer) Supply() error {
	g.context.Log.BeginStep("Supplying Groovy")

	// Install Groovy runtime
	dep, err := g.context.Manifest.DefaultVersion("groovy")
	if err != nil {
		g.context.Log.Warning("Unable to determine default Groovy version")
		// Fallback version
		dep.Name = "groovy"
		dep.Version = "4.0.0"
	}

	groovyDir := filepath.Join(g.context.Stager.DepDir(), "groovy")
	if err := g.context.Installer.InstallDependency(dep, groovyDir); err != nil {
		return fmt.Errorf("failed to install Groovy: %w", err)
	}

	g.context.Log.Info("Installed Groovy version %s", dep.Version)

	// Set GROOVY_HOME
	if err := g.context.Stager.WriteEnvFile("GROOVY_HOME", groovyDir); err != nil {
		return fmt.Errorf("failed to set GROOVY_HOME: %w", err)
	}

	// Add Groovy bin to PATH
	groovyBin := filepath.Join(groovyDir, "bin")
	if err := g.context.Stager.AddBinDependencyLink(groovyBin, "groovy"); err != nil {
		g.context.Log.Warning("Could not link groovy binary: %s", err.Error())
	}

	// Install JVMKill agent
	if err := g.installJVMKillAgent(); err != nil {
		g.context.Log.Warning("Could not install JVMKill agent: %s", err.Error())
	}

	return nil
}

// installJVMKillAgent installs the JVMKill agent
func (g *GroovyContainer) installJVMKillAgent() error {
	dep, err := g.context.Manifest.DefaultVersion("jvmkill")
	if err != nil {
		return err
	}

	jvmkillPath := filepath.Join(g.context.Stager.DepDir(), "jvmkill")
	if err := g.context.Installer.InstallDependency(dep, jvmkillPath); err != nil {
		return fmt.Errorf("failed to install JVMKill: %w", err)
	}

	g.context.Log.Info("Installed JVMKill agent version %s", dep.Version)
	return nil
}

// Finalize performs final Groovy configuration
func (g *GroovyContainer) Finalize() error {
	g.context.Log.BeginStep("Finalizing Groovy")

	// Configure JAVA_OPTS for Groovy
	javaOpts := []string{
		"-Djava.io.tmpdir=$TMPDIR",
		"-XX:+ExitOnOutOfMemoryError",
	}

	// Add JVMKill agent if available
	jvmkillSO := filepath.Join(g.context.Stager.DepDir(), "jvmkill", "jvmkill.so")
	if _, err := os.Stat(jvmkillSO); err == nil {
		javaOpts = append(javaOpts, fmt.Sprintf("-agentpath:%s", jvmkillSO))
	}

	// Write JAVA_OPTS
	if err := g.context.Stager.WriteEnvFile("JAVA_OPTS", strings.Join(javaOpts, " ")); err != nil {
		return fmt.Errorf("failed to write JAVA_OPTS: %w", err)
	}

	return nil
}

// Release returns the Groovy startup command
func (g *GroovyContainer) Release() (string, error) {
	// Determine which script to run
	var mainScript string

	// Check for GROOVY_SCRIPT environment variable
	mainScript = os.Getenv("GROOVY_SCRIPT")

	if mainScript == "" && len(g.groovyScripts) > 0 {
		// Use the first Groovy script found
		mainScript = filepath.Base(g.groovyScripts[0])
	}

	if mainScript == "" {
		return "", fmt.Errorf("no Groovy script specified (set GROOVY_SCRIPT)")
	}

	cmd := fmt.Sprintf("$GROOVY_HOME/bin/groovy $JAVA_OPTS %s", mainScript)
	return cmd, nil
}
