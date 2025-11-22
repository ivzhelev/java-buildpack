package containers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// JavaMainContainer handles standalone JAR applications with a main class
type JavaMainContainer struct {
	context   *Context
	mainClass string
	jarFile   string
}

// NewJavaMainContainer creates a new Java Main container
func NewJavaMainContainer(ctx *Context) *JavaMainContainer {
	return &JavaMainContainer{
		context: ctx,
	}
}

// Detect checks if this is a Java Main application
func (j *JavaMainContainer) Detect() (string, error) {
	buildDir := j.context.Stager.BuildDir()

	// Look for JAR files with Main-Class manifest
	mainClass, jarFile := j.findMainClass(buildDir)
	if mainClass != "" {
		j.mainClass = mainClass
		j.jarFile = jarFile
		j.context.Log.Debug("Detected Java Main application: %s (main: %s)", jarFile, mainClass)
		return "Java Main", nil
	}

	// Check for META-INF/MANIFEST.MF with Main-Class
	manifestPath := filepath.Join(buildDir, "META-INF", "MANIFEST.MF")
	if _, err := os.Stat(manifestPath); err == nil {
		// Read manifest for Main-Class
		if mainClass := j.readMainClassFromManifest(manifestPath); mainClass != "" {
			j.mainClass = mainClass
			j.context.Log.Debug("Detected Java Main application via MANIFEST.MF: %s", mainClass)
			return "Java Main", nil
		}
	}

	// Check for compiled .class files
	classFiles, err := filepath.Glob(filepath.Join(buildDir, "*.class"))
	if err == nil && len(classFiles) > 0 {
		j.context.Log.Debug("Detected compiled Java classes")
		return "Java Main", nil
	}

	return "", nil
}

// findMainClass searches for a JAR with a Main-Class manifest entry
func (j *JavaMainContainer) findMainClass(buildDir string) (string, string) {
	entries, err := os.ReadDir(buildDir)
	if err != nil {
		return "", ""
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if strings.HasSuffix(name, ".jar") {
			// TODO: In full implementation, extract and read MANIFEST.MF
			// For now, assume any JAR could be a main JAR
			return "Main", name
		}
	}

	return "", ""
}

// readMainClassFromManifest reads the Main-Class from a manifest file
func (j *JavaMainContainer) readMainClassFromManifest(manifestPath string) string {
	// TODO: In full implementation, parse MANIFEST.MF properly
	// For now, return empty to trigger alternative detection
	return ""
}

// Supply installs Java Main dependencies
func (j *JavaMainContainer) Supply() error {
	j.context.Log.BeginStep("Supplying Java Main")

	// For Java Main apps, we need to:
	// 1. Ensure all JARs are available
	// 2. Set up classpath
	// 3. Install support utilities

	// Install JVMKill agent
	if err := j.installJVMKillAgent(); err != nil {
		j.context.Log.Warning("Could not install JVMKill agent: %s", err.Error())
	}

	return nil
}

// installJVMKillAgent installs the JVMKill agent
func (j *JavaMainContainer) installJVMKillAgent() error {
	dep, err := j.context.Manifest.DefaultVersion("jvmkill")
	if err != nil {
		return err
	}

	jvmkillPath := filepath.Join(j.context.Stager.DepDir(), "jvmkill")
	if err := j.context.Installer.InstallDependency(dep, jvmkillPath); err != nil {
		return fmt.Errorf("failed to install JVMKill: %w", err)
	}

	j.context.Log.Info("Installed JVMKill agent version %s", dep.Version)
	return nil
}

// Finalize performs final Java Main configuration
func (j *JavaMainContainer) Finalize() error {
	j.context.Log.BeginStep("Finalizing Java Main")

	// Build classpath
	classpath, err := j.buildClasspath()
	if err != nil {
		return fmt.Errorf("failed to build classpath: %w", err)
	}

	// Write CLASSPATH environment variable
	if err := j.context.Stager.WriteEnvFile("CLASSPATH", classpath); err != nil {
		return fmt.Errorf("failed to write CLASSPATH: %w", err)
	}

	// Configure JAVA_OPTS
	javaOpts := []string{
		"-Djava.io.tmpdir=$TMPDIR",
		"-XX:+ExitOnOutOfMemoryError",
	}

	// Add JVMKill agent if available
	jvmkillSO := filepath.Join(j.context.Stager.DepDir(), "jvmkill", "jvmkill.so")
	if _, err := os.Stat(jvmkillSO); err == nil {
		javaOpts = append(javaOpts, fmt.Sprintf("-agentpath:%s", jvmkillSO))
	}

	// Write JAVA_OPTS
	if err := j.context.Stager.WriteEnvFile("JAVA_OPTS", strings.Join(javaOpts, " ")); err != nil {
		return fmt.Errorf("failed to write JAVA_OPTS: %w", err)
	}

	return nil
}

// buildClasspath builds the classpath for the application
func (j *JavaMainContainer) buildClasspath() (string, error) {
	buildDir := j.context.Stager.BuildDir()

	var classpathEntries []string

	// Add current directory
	classpathEntries = append(classpathEntries, ".")

	// Add all JARs in the build directory
	jarFiles, err := filepath.Glob(filepath.Join(buildDir, "*.jar"))
	if err == nil {
		for _, jar := range jarFiles {
			classpathEntries = append(classpathEntries, filepath.Base(jar))
		}
	}

	// Add lib directory if it exists
	libDir := filepath.Join(buildDir, "lib")
	if _, err := os.Stat(libDir); err == nil {
		libJars, err := filepath.Glob(filepath.Join(libDir, "*.jar"))
		if err == nil {
			for _, jar := range libJars {
				relPath, _ := filepath.Rel(buildDir, jar)
				classpathEntries = append(classpathEntries, relPath)
			}
		}
	}

	return strings.Join(classpathEntries, ":"), nil
}

// Release returns the Java Main startup command
func (j *JavaMainContainer) Release() (string, error) {
	// Determine the main class to run
	mainClass := j.mainClass
	if mainClass == "" {
		// Try to detect from environment or configuration
		mainClass = os.Getenv("JAVA_MAIN_CLASS")
		if mainClass == "" {
			return "", fmt.Errorf("no main class specified (set JAVA_MAIN_CLASS)")
		}
	}

	var cmd string
	if j.jarFile != "" {
		// Run from JAR
		cmd = fmt.Sprintf("$JAVA_HOME/bin/java $JAVA_OPTS -jar %s", j.jarFile)
	} else {
		// Run with classpath and main class
		cmd = fmt.Sprintf("$JAVA_HOME/bin/java $JAVA_OPTS -cp $CLASSPATH %s", mainClass)
	}

	return cmd, nil
}
