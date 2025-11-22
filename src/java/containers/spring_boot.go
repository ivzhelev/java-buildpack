package containers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SpringBootContainer handles Spring Boot JAR applications
type SpringBootContainer struct {
	context *Context
	jarFile string
}

// NewSpringBootContainer creates a new Spring Boot container
func NewSpringBootContainer(ctx *Context) *SpringBootContainer {
	return &SpringBootContainer{
		context: ctx,
	}
}

// Detect checks if this is a Spring Boot application
func (s *SpringBootContainer) Detect() (string, error) {
	buildDir := s.context.Stager.BuildDir()

	// Check for BOOT-INF directory (exploded Spring Boot JAR)
	bootInf := filepath.Join(buildDir, "BOOT-INF")
	if _, err := os.Stat(bootInf); err == nil {
		s.context.Log.Debug("Detected Spring Boot application via BOOT-INF directory")
		return "Spring Boot", nil
	}

	// Check for Spring Boot JAR
	jarFile, err := s.findSpringBootJar(buildDir)
	if err == nil && jarFile != "" {
		s.jarFile = jarFile
		s.context.Log.Debug("Detected Spring Boot JAR: %s", jarFile)
		return "Spring Boot", nil
	}

	return "", nil
}

// findSpringBootJar looks for a Spring Boot JAR in the build directory
func (s *SpringBootContainer) findSpringBootJar(buildDir string) (string, error) {
	entries, err := os.ReadDir(buildDir)
	if err != nil {
		return "", err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if strings.HasSuffix(name, ".jar") {
			// Check if JAR has Spring Boot manifest
			jarPath := filepath.Join(buildDir, name)
			if s.isSpringBootJar(jarPath) {
				return name, nil
			}
		}
	}

	return "", nil
}

// isSpringBootJar checks if a JAR is a Spring Boot JAR
func (s *SpringBootContainer) isSpringBootJar(jarPath string) bool {
	// TODO: In full implementation, we'd extract and check MANIFEST.MF
	// For now, check file name patterns
	name := filepath.Base(jarPath)
	return strings.Contains(name, "spring") ||
		strings.Contains(name, "boot") ||
		strings.Contains(name, "BOOT-INF")
}

// Supply installs Spring Boot dependencies
func (s *SpringBootContainer) Supply() error {
	s.context.Log.BeginStep("Supplying Spring Boot")

	// For Spring Boot, most dependencies are already in the JAR
	// JRE installation (including JVMKill and Memory Calculator) is handled by the JRE provider
	// No additional installation needed for Spring Boot

	return nil
}

// Finalize performs final Spring Boot configuration
func (s *SpringBootContainer) Finalize() error {
	s.context.Log.BeginStep("Finalizing Spring Boot")

	// Read existing JAVA_OPTS (set by JRE finalize phase)
	envFile := filepath.Join(s.context.Stager.DepDir(), "env", "JAVA_OPTS")
	var existingOpts string
	if data, err := os.ReadFile(envFile); err == nil {
		existingOpts = strings.TrimSpace(string(data))
	}

	// Configure additional JAVA_OPTS for Spring Boot
	additionalOpts := []string{
		"-Djava.io.tmpdir=$TMPDIR",
		"-XX:+ExitOnOutOfMemoryError",
	}

	// Combine existing opts with additional opts
	var finalOpts string
	if existingOpts != "" {
		finalOpts = existingOpts + " " + strings.Join(additionalOpts, " ")
	} else {
		finalOpts = strings.Join(additionalOpts, " ")
	}

	// Write combined JAVA_OPTS
	if err := s.context.Stager.WriteEnvFile("JAVA_OPTS", finalOpts); err != nil {
		return fmt.Errorf("failed to write JAVA_OPTS: %w", err)
	}

	return nil
}

// Release returns the Spring Boot startup command
func (s *SpringBootContainer) Release() (string, error) {
	buildDir := s.context.Stager.BuildDir()

	// Check if we have an exploded JAR (BOOT-INF directory)
	bootInf := filepath.Join(buildDir, "BOOT-INF")
	if _, err := os.Stat(bootInf); err == nil {
		// Exploded JAR - use Spring Boot's launcher
		return "$JAVA_HOME/bin/java $JAVA_OPTS -cp . org.springframework.boot.loader.JarLauncher", nil
	}

	// Find the Spring Boot JAR
	jarFile := s.jarFile
	if jarFile == "" {
		jar, err := s.findSpringBootJar(buildDir)
		if err != nil || jar == "" {
			return "", fmt.Errorf("no Spring Boot JAR found")
		}
		jarFile = jar
	}

	cmd := fmt.Sprintf("$JAVA_HOME/bin/java $JAVA_OPTS -jar %s", jarFile)
	return cmd, nil
}
