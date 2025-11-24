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

// IntroscopeAgentFramework represents the CA APM Introscope agent framework
type IntroscopeAgentFramework struct {
	context   *Context
	agentPath string
}

// NewIntroscopeAgentFramework creates a new Introscope agent framework instance
func NewIntroscopeAgentFramework(ctx *Context) *IntroscopeAgentFramework {
	return &IntroscopeAgentFramework{context: ctx}
}

// Detect checks if Introscope agent should be enabled
func (i *IntroscopeAgentFramework) Detect() (string, error) {
	// Check for introscope service binding
	if i.hasServiceBinding() {
		i.context.Log.Debug("Introscope agent framework detected via service binding")
		return "introscope-agent", nil
	}

	i.context.Log.Debug("Introscope agent: no service binding found")
	return "", nil
}

// Supply downloads and installs the Introscope agent
func (i *IntroscopeAgentFramework) Supply() error {
	i.context.Log.BeginStep("Installing Introscope agent")

	// Get dependency from manifest
	dep, err := i.context.Manifest.DefaultVersion("introscope-agent")
	if err != nil {
		i.context.Log.Warning("Unable to find Introscope agent in manifest: %s", err)
		return nil // Non-blocking
	}

	// Install the agent
	agentDir := filepath.Join(i.context.Stager.DepDir(), "introscope_agent")
	if err := i.context.Installer.InstallDependency(dep, agentDir); err != nil {
		i.context.Log.Warning("Failed to install Introscope agent: %s", err)
		return nil // Non-blocking
	}

	// Find the installed agent JAR
	agentPattern := filepath.Join(agentDir, "Agent.jar")
	if _, err := os.Stat(agentPattern); err != nil {
		i.context.Log.Warning("Introscope Agent.jar not found after installation")
		return nil
	}
	i.agentPath = agentPattern

	i.context.Log.Info("Introscope agent %s installed", dep.Version)
	return nil
}

// Finalize configures the Introscope agent
func (i *IntroscopeAgentFramework) Finalize() error {
	if i.agentPath == "" {
		return nil
	}

	i.context.Log.BeginStep("Configuring Introscope agent")

	// Get credentials from service binding
	credentials := i.getCredentials()

	// Add javaagent to JAVA_OPTS
	javaagentOpt := fmt.Sprintf("-javaagent:%s", i.agentPath)
	if err := i.appendToJavaOpts(javaagentOpt); err != nil {
		i.context.Log.Warning("Failed to add Introscope agent to JAVA_OPTS: %s", err)
		return nil
	}

	// Configure agent name (default to application name)
	agentName := credentials.AgentName
	if agentName == "" {
		agentName = i.getApplicationName()
	}
	if agentName != "" {
		nameOpt := fmt.Sprintf("-Dcom.wily.introscope.agentProfile.agent.name=%s", agentName)
		if err := i.appendToJavaOpts(nameOpt); err != nil {
			i.context.Log.Warning("Failed to set agent name: %s", err)
		}
	}

	// Configure Enterprise Manager host
	if credentials.EMHost != "" {
		hostOpt := fmt.Sprintf("-Dcom.wily.introscope.agentProfile.agent.enterpriseManager.host=%s", credentials.EMHost)
		if err := i.appendToJavaOpts(hostOpt); err != nil {
			i.context.Log.Warning("Failed to set EM host: %s", err)
		}
	}

	// Configure Enterprise Manager port
	if credentials.EMPort != "" {
		portOpt := fmt.Sprintf("-Dcom.wily.introscope.agentProfile.agent.enterpriseManager.port=%s", credentials.EMPort)
		if err := i.appendToJavaOpts(portOpt); err != nil {
			i.context.Log.Warning("Failed to set EM port: %s", err)
		}
	}

	i.context.Log.Info("Introscope agent configured")
	return nil
}

// hasServiceBinding checks if there's an introscope service binding
func (i *IntroscopeAgentFramework) hasServiceBinding() bool {
	vcapServices := os.Getenv("VCAP_SERVICES")
	if vcapServices == "" {
		return false
	}

	var services map[string][]map[string]interface{}
	if err := json.Unmarshal([]byte(vcapServices), &services); err != nil {
		return false
	}

	// Check for introscope service
	serviceNames := []string{
		"introscope",
		"ca-apm",
		"ca-wily",
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
						tagLower := strings.ToLower(tagStr)
						if strings.Contains(tagLower, "introscope") ||
							strings.Contains(tagLower, "ca-apm") ||
							strings.Contains(tagLower, "wily") {
							return true
						}
					}
				}
			}
		}
	}

	return false
}

// IntroscopeCredentials holds Introscope agent credentials
type IntroscopeCredentials struct {
	AgentName string
	EMHost    string
	EMPort    string
}

// getCredentials retrieves Introscope credentials from service binding
func (i *IntroscopeAgentFramework) getCredentials() IntroscopeCredentials {
	creds := IntroscopeCredentials{}

	vcapServices := os.Getenv("VCAP_SERVICES")
	if vcapServices == "" {
		return creds
	}

	var services map[string][]map[string]interface{}
	if err := json.Unmarshal([]byte(vcapServices), &services); err != nil {
		return creds
	}

	// Look for introscope service
	serviceNames := []string{
		"introscope",
		"ca-apm",
		"ca-wily",
		"user-provided",
	}

	for _, serviceName := range serviceNames {
		if serviceList, ok := services[serviceName]; ok {
			for _, service := range serviceList {
				if credentials, ok := service["credentials"].(map[string]interface{}); ok {
					// Get agent name
					if agentName, ok := credentials["agent_name"].(string); ok {
						creds.AgentName = agentName
					} else if agentName, ok := credentials["agentName"].(string); ok {
						creds.AgentName = agentName
					}

					// Get EM host
					if emHost, ok := credentials["em_host"].(string); ok {
						creds.EMHost = emHost
					} else if emHost, ok := credentials["emHost"].(string); ok {
						creds.EMHost = emHost
					}

					// Get EM port
					if emPort, ok := credentials["em_port"].(string); ok {
						creds.EMPort = emPort
					} else if emPort, ok := credentials["emPort"].(string); ok {
						creds.EMPort = emPort
					} else if emPort, ok := credentials["em_port"].(float64); ok {
						creds.EMPort = fmt.Sprintf("%.0f", emPort)
					} else if emPort, ok := credentials["emPort"].(float64); ok {
						creds.EMPort = fmt.Sprintf("%.0f", emPort)
					}

					if creds.EMHost != "" {
						return creds
					}
				}
			}
		}
	}

	return creds
}

// getApplicationName returns the application name from VCAP_APPLICATION
func (i *IntroscopeAgentFramework) getApplicationName() string {
	vcapApp := os.Getenv("VCAP_APPLICATION")
	if vcapApp == "" {
		return ""
	}

	var appData map[string]interface{}
	if err := json.Unmarshal([]byte(vcapApp), &appData); err != nil {
		return ""
	}

	if name, ok := appData["application_name"].(string); ok {
		return name
	}

	return ""
}

// appendToJavaOpts appends a value to JAVA_OPTS
func (i *IntroscopeAgentFramework) appendToJavaOpts(value string) error {
	javaOptsFile := filepath.Join(i.context.Stager.DepDir(), "env", "JAVA_OPTS")

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
	return i.context.Stager.WriteEnvFile(javaOptsFile, existingOpts)
}
