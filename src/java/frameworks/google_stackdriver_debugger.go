// Cloud Foundry Java Buildpack
// Copyright 2013-2021 the original author or authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package frameworks

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GoogleStackdriverDebuggerFramework represents the Google Stackdriver Debugger framework
type GoogleStackdriverDebuggerFramework struct {
	context   *Context
	agentPath string
}

// NewGoogleStackdriverDebuggerFramework creates a new Google Stackdriver Debugger framework instance
func NewGoogleStackdriverDebuggerFramework(ctx *Context) *GoogleStackdriverDebuggerFramework {
	return &GoogleStackdriverDebuggerFramework{context: ctx}
}

// Detect checks if Google Stackdriver Debugger should be enabled
func (g *GoogleStackdriverDebuggerFramework) Detect() (string, error) {
	// Check for google-stackdriver-debugger service binding
	if g.hasServiceBinding() {
		g.context.Log.Debug("Google Stackdriver Debugger framework detected via service binding")
		return "google-stackdriver-debugger", nil
	}

	// Check for GOOGLE_APPLICATION_CREDENTIALS
	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") != "" {
		g.context.Log.Debug("Google Stackdriver Debugger framework detected via GOOGLE_APPLICATION_CREDENTIALS")
		return "google-stackdriver-debugger", nil
	}

	g.context.Log.Debug("Google Stackdriver Debugger: no service binding found")
	return "", nil
}

// Supply downloads and installs the Google Stackdriver Debugger
func (g *GoogleStackdriverDebuggerFramework) Supply() error {
	g.context.Log.BeginStep("Installing Google Stackdriver Debugger")

	// Get dependency from manifest
	dep, err := g.context.Manifest.DefaultVersion("google-stackdriver-debugger")
	if err != nil {
		g.context.Log.Warning("Unable to find Google Stackdriver Debugger in manifest: %s", err)
		return nil // Non-blocking
	}

	// Install the debugger
	debuggerDir := filepath.Join(g.context.Stager.DepDir(), "google_stackdriver_debugger")
	if err := g.context.Installer.InstallDependency(dep, debuggerDir); err != nil {
		g.context.Log.Warning("Failed to install Google Stackdriver Debugger: %s", err)
		return nil // Non-blocking
	}

	// Find the installed agent (native library)
	agentPattern := filepath.Join(debuggerDir, "cdbg_java_agent.so")
	if _, err := os.Stat(agentPattern); err != nil {
		g.context.Log.Warning("Google Stackdriver Debugger agent not found after installation")
		return nil
	}
	g.agentPath = agentPattern

	g.context.Log.Info("Google Stackdriver Debugger %s installed", dep.Version)
	return nil
}

// Finalize configures the Google Stackdriver Debugger
func (g *GoogleStackdriverDebuggerFramework) Finalize() error {
	if g.agentPath == "" {
		return nil
	}

	g.context.Log.BeginStep("Configuring Google Stackdriver Debugger")

	// Get credentials
	credentials := g.getCredentials()

	// Add agentpath to JAVA_OPTS
	agentOpt := fmt.Sprintf("-agentpath:%s", g.agentPath)

	// Add project ID if available
	if credentials.ProjectID != "" {
		agentOpt += fmt.Sprintf("=-Dcom.google.cdbg.module=%s", credentials.ProjectID)
	}

	if err := g.appendToJavaOpts(agentOpt); err != nil {
		g.context.Log.Warning("Failed to add Google Stackdriver Debugger to JAVA_OPTS: %s", err)
		return nil
	}

	// Set application version
	if appVersion := g.getApplicationVersion(); appVersion != "" {
		versionOpt := fmt.Sprintf("-Dcom.google.cdbg.version=%s", appVersion)
		if err := g.appendToJavaOpts(versionOpt); err != nil {
			g.context.Log.Warning("Failed to set debugger version: %s", err)
		}
	}

	g.context.Log.Info("Google Stackdriver Debugger configured")
	return nil
}

// hasServiceBinding checks if there's a google-stackdriver-debugger service binding
func (g *GoogleStackdriverDebuggerFramework) hasServiceBinding() bool {
	vcapServices := os.Getenv("VCAP_SERVICES")
	if vcapServices == "" {
		return false
	}

	var services map[string][]map[string]interface{}
	if err := json.Unmarshal([]byte(vcapServices), &services); err != nil {
		return false
	}

	// Check for Google Stackdriver Debugger service
	serviceNames := []string{
		"google-stackdriver-debugger",
		"stackdriver-debugger",
	}

	for _, serviceName := range serviceNames {
		if serviceList, ok := services[serviceName]; ok && len(serviceList) > 0 {
			return true
		}
	}

	// Check user-provided services
	if userProvided, ok := services["user-provided"]; ok {
		for _, service := range userProvided {
			if tags, ok := service["tags"].([]interface{}); ok {
				for _, tag := range tags {
					if tagStr, ok := tag.(string); ok {
						if strings.Contains(strings.ToLower(tagStr), "stackdriver-debugger") {
							return true
						}
					}
				}
			}
		}
	}

	return false
}

// GoogleCredentials holds Google Cloud credentials
type GoogleCredentials struct {
	ProjectID string
}

// getCredentials retrieves Google Cloud credentials
func (g *GoogleStackdriverDebuggerFramework) getCredentials() GoogleCredentials {
	creds := GoogleCredentials{}

	vcapServices := os.Getenv("VCAP_SERVICES")
	if vcapServices == "" {
		return creds
	}

	var services map[string][]map[string]interface{}
	if err := json.Unmarshal([]byte(vcapServices), &services); err != nil {
		return creds
	}

	// Look for Google service
	serviceNames := []string{
		"google-stackdriver-debugger",
		"stackdriver-debugger",
		"user-provided",
	}

	for _, serviceName := range serviceNames {
		if serviceList, ok := services[serviceName]; ok {
			for _, service := range serviceList {
				if credentials, ok := service["credentials"].(map[string]interface{}); ok {
					if projectID, ok := credentials["ProjectId"].(string); ok {
						creds.ProjectID = projectID
						return creds
					}
					if projectID, ok := credentials["project_id"].(string); ok {
						creds.ProjectID = projectID
						return creds
					}
				}
			}
		}
	}

	return creds
}

// getApplicationVersion returns the application version
func (g *GoogleStackdriverDebuggerFramework) getApplicationVersion() string {
	vcapApp := os.Getenv("VCAP_APPLICATION")
	if vcapApp == "" {
		return ""
	}

	var appData map[string]interface{}
	if err := json.Unmarshal([]byte(vcapApp), &appData); err != nil {
		return ""
	}

	if version, ok := appData["application_version"].(string); ok {
		return version
	}

	return ""
}

// appendToJavaOpts appends a value to JAVA_OPTS
func (g *GoogleStackdriverDebuggerFramework) appendToJavaOpts(value string) error {
	javaOptsFile := filepath.Join(g.context.Stager.DepDir(), "env", "JAVA_OPTS")

	// Read existing JAVA_OPTS
	var existingOpts string
	if data, err := os.ReadFile(javaOptsFile); err == nil {
		existingOpts = string(data)
	}

	// Append new value
	if existingOpts != "" {
		existingOpts += " "
	}
	existingOpts += value

	// Write back
	return g.context.Stager.WriteEnvFile(javaOptsFile, existingOpts)
}
