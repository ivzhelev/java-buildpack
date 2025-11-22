package jres

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudfoundry/libbuildpack"
)

// JRE represents a Java Runtime Environment provider
type JRE interface {
	// Name returns the name of this JRE provider (e.g., "OpenJDK", "Zulu")
	Name() string

	// Detect returns true if this JRE should be used
	Detect() (bool, error)

	// Supply installs the JRE and its components (memory calculator, jvmkill)
	Supply() error

	// Finalize performs any final JRE configuration
	Finalize() error

	// JavaHome returns the path to JAVA_HOME
	JavaHome() string

	// Version returns the installed JRE version
	Version() string
}

// Context holds shared dependencies for JRE providers
type Context struct {
	Stager    *libbuildpack.Stager
	Manifest  *libbuildpack.Manifest
	Installer *libbuildpack.Installer
	Log       *libbuildpack.Logger
	Command   *libbuildpack.Command
}

// Registry manages multiple JRE providers
type Registry struct {
	ctx       *Context
	providers []JRE
}

// NewRegistry creates a new JRE registry
func NewRegistry(ctx *Context) *Registry {
	return &Registry{
		ctx:       ctx,
		providers: []JRE{},
	}
}

// Register adds a JRE provider to the registry
func (r *Registry) Register(jre JRE) {
	r.providers = append(r.providers, jre)
}

// Detect finds the first JRE provider that should be used
// Returns the JRE, its name, and any error
func (r *Registry) Detect() (JRE, string, error) {
	for _, jre := range r.providers {
		detected, err := jre.Detect()
		if err != nil {
			r.ctx.Log.Warning("Error detecting JRE %s: %s", jre.Name(), err.Error())
			continue
		}
		if detected {
			return jre, jre.Name(), nil
		}
	}
	return nil, "", nil
}

// Component represents a JRE component (memory calculator, jvmkill, etc.)
type Component interface {
	// Name returns the component name
	Name() string

	// Supply installs the component
	Supply() error

	// Finalize performs final configuration
	Finalize() error
}

// BaseComponent provides common functionality for JRE components
type BaseComponent struct {
	Ctx         *Context
	JREDir      string
	JREVersion  string
	ComponentID string
}

// Memory calculator constants
const (
	DefaultStackThreads = 250
	DefaultHeadroom     = 0
	Java9ClassCount     = 42215 // Classes in Java 9+ JRE
)

// Helper functions

// DetectJREByEnv checks environment variables for JRE selection
// Supports JBP_CONFIG_OPEN_JDK_JRE, etc.
func DetectJREByEnv(jreName string) bool {
	envKey := fmt.Sprintf("JBP_CONFIG_%s", strings.ToUpper(strings.ReplaceAll(jreName, "-", "_")))
	return os.Getenv(envKey) != ""
}

// GetJREVersion gets the desired JRE version from environment or uses default
func GetJREVersion(ctx *Context, jreName string) (libbuildpack.Dependency, error) {
	// Check for version override in environment
	envKey := fmt.Sprintf("JBP_CONFIG_%s", strings.ToUpper(strings.ReplaceAll(jreName, "-", "_")))
	if envVal := os.Getenv(envKey); envVal != "" {
		// Parse version from env (e.g., '{jre: {version: 11.+}}')
		// For now, simplified
		ctx.Log.Debug("JRE version override from environment: %s", envVal)
	}

	// Get default version from manifest
	dep, err := ctx.Manifest.DefaultVersion(jreName)
	if err != nil {
		return libbuildpack.Dependency{}, err
	}

	return dep, nil
}

// SetupJavaHome sets JAVA_HOME and related environment variables
func SetupJavaHome(ctx *Context, javaHome string) error {
	// Find actual JRE directory (usually jdk-* or jre-* subdirectory)
	entries, err := os.ReadDir(javaHome)
	if err != nil {
		return fmt.Errorf("failed to read JRE directory: %w", err)
	}

	// Look for jdk-* or jre-* subdirectory
	var actualJavaHome string
	for _, entry := range entries {
		if entry.IsDir() && (strings.HasPrefix(entry.Name(), "jdk") || strings.HasPrefix(entry.Name(), "jre")) {
			actualJavaHome = filepath.Join(javaHome, entry.Name())
			break
		}
	}

	// If no subdirectory found, use the javaHome directly
	if actualJavaHome == "" {
		actualJavaHome = javaHome
	}

	// Write environment variables to profile.d
	envScript := filepath.Join(ctx.Stager.DepDir(), "profile.d", "java.sh")
	if err := os.MkdirAll(filepath.Dir(envScript), 0755); err != nil {
		return fmt.Errorf("failed to create profile.d directory: %w", err)
	}

	envContent := fmt.Sprintf(`export JAVA_HOME=%s
export JRE_HOME=%s
export PATH=$JAVA_HOME/bin:$PATH
`, actualJavaHome, actualJavaHome)

	if err := os.WriteFile(envScript, []byte(envContent), 0755); err != nil {
		return fmt.Errorf("failed to write java.sh: %w", err)
	}

	// Also set for current process
	os.Setenv("JAVA_HOME", actualJavaHome)
	os.Setenv("JRE_HOME", actualJavaHome)
	os.Setenv("PATH", filepath.Join(actualJavaHome, "bin")+":"+os.Getenv("PATH"))

	ctx.Log.Info("Set JAVA_HOME to %s", actualJavaHome)

	return nil
}

// DetermineJavaVersion determines the major Java version from the installed JRE
func DetermineJavaVersion(javaHome string) (int, error) {
	// Try to read release file
	releaseFile := filepath.Join(javaHome, "release")
	if data, err := os.ReadFile(releaseFile); err == nil {
		// Parse JAVA_VERSION="1.8.0_422" or JAVA_VERSION="17.0.13"
		content := string(data)
		for _, line := range strings.Split(content, "\n") {
			if strings.HasPrefix(line, "JAVA_VERSION=") {
				version := strings.Trim(strings.TrimPrefix(line, "JAVA_VERSION="), "\"")
				// Parse major version
				if strings.HasPrefix(version, "1.8") {
					return 8, nil
				}
				// For Java 9+, major version is the first number
				parts := strings.Split(version, ".")
				if len(parts) > 0 {
					var major int
					fmt.Sscanf(parts[0], "%d", &major)
					return major, nil
				}
			}
		}
	}

	// Default to 17 if we can't determine
	return 17, nil
}

// WriteJavaOpts writes JAVA_OPTS to an environment file
func WriteJavaOpts(ctx *Context, opts string) error {
	envFile := filepath.Join(ctx.Stager.DepDir(), "env", "JAVA_OPTS")
	if err := os.MkdirAll(filepath.Dir(envFile), 0755); err != nil {
		return fmt.Errorf("failed to create env directory: %w", err)
	}

	// Append to existing JAVA_OPTS if file exists
	var content string
	if existing, err := os.ReadFile(envFile); err == nil {
		content = string(existing) + " " + opts
	} else {
		content = opts
	}

	if err := os.WriteFile(envFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write JAVA_OPTS: %w", err)
	}

	return nil
}
