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
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// CheckmarxIASTAgentFramework represents the Checkmarx IAST agent framework
type CheckmarxIASTAgentFramework struct {
	context *Context
	jarPath string
}

// NewCheckmarxIASTAgentFramework creates a new Checkmarx IAST agent framework instance
func NewCheckmarxIASTAgentFramework(ctx *Context) *CheckmarxIASTAgentFramework {
	return &CheckmarxIASTAgentFramework{context: ctx}
}

// Detect checks if Checkmarx IAST agent should be enabled
func (c *CheckmarxIASTAgentFramework) Detect() (string, error) {
	// Check for checkmarx-iast service binding
	if c.hasServiceBinding() {
		c.context.Log.Debug("Checkmarx IAST agent framework detected via service binding")
		return "checkmarx-iast-agent", nil
	}

	c.context.Log.Debug("Checkmarx IAST agent: no service binding found")
	return "", nil
}

// Supply downloads and installs the Checkmarx IAST agent
func (c *CheckmarxIASTAgentFramework) Supply() error {
	c.context.Log.BeginStep("Installing Checkmarx IAST agent")

	// Get credentials from service binding
	credentials := c.getCredentials()
	if credentials.URL == "" {
		c.context.Log.Warning("Checkmarx IAST agent URL not found in service binding")
		return nil // Non-blocking
	}

	// Download the agent from the URL provided in service credentials
	agentDir := filepath.Join(c.context.Stager.DepDir(), "checkmarx_iast_agent")
	if err := os.MkdirAll(agentDir, 0755); err != nil {
		c.context.Log.Warning("Failed to create Checkmarx IAST agent directory: %s", err)
		return nil
	}

	jarPath := filepath.Join(agentDir, "cx-agent.jar")
	if err := c.downloadAgent(credentials.URL, jarPath); err != nil {
		c.context.Log.Warning("Failed to download Checkmarx IAST agent: %s", err)
		return nil
	}

	c.jarPath = jarPath
	c.context.Log.Info("Checkmarx IAST agent installed from %s", credentials.URL)
	return nil
}

// Finalize configures the Checkmarx IAST agent
func (c *CheckmarxIASTAgentFramework) Finalize() error {
	if c.jarPath == "" {
		return nil
	}

	c.context.Log.BeginStep("Configuring Checkmarx IAST agent")

	// Get credentials
	credentials := c.getCredentials()

	// Add javaagent to JAVA_OPTS
	javaagentOpt := fmt.Sprintf("-javaagent:%s", c.jarPath)
	if err := c.appendToJavaOpts(javaagentOpt); err != nil {
		c.context.Log.Warning("Failed to add Checkmarx IAST agent to JAVA_OPTS: %s", err)
		return nil
	}

	// Set Checkmarx manager URL if available
	if credentials.ManagerURL != "" {
		managerOpt := fmt.Sprintf("-Dcheckmarx.manager.url=%s", credentials.ManagerURL)
		if err := c.appendToJavaOpts(managerOpt); err != nil {
			c.context.Log.Warning("Failed to set Checkmarx manager URL: %s", err)
		}
	}

	// Set API key if available
	if credentials.APIKey != "" {
		apiKeyOpt := fmt.Sprintf("-Dcheckmarx.api.key=%s", credentials.APIKey)
		if err := c.appendToJavaOpts(apiKeyOpt); err != nil {
			c.context.Log.Warning("Failed to set Checkmarx API key: %s", err)
		}
	}

	c.context.Log.Info("Checkmarx IAST agent configured")
	return nil
}

// hasServiceBinding checks if there's a checkmarx-iast service binding
func (c *CheckmarxIASTAgentFramework) hasServiceBinding() bool {
	vcapServices := os.Getenv("VCAP_SERVICES")
	if vcapServices == "" {
		return false
	}

	var services map[string][]map[string]interface{}
	if err := json.Unmarshal([]byte(vcapServices), &services); err != nil {
		return false
	}

	// Check for checkmarx-iast service
	serviceNames := []string{
		"checkmarx-iast",
		"checkmarx",
	}

	for _, serviceName := range serviceNames {
		if serviceList, ok := services[serviceName]; ok && len(serviceList) > 0 {
			return true
		}
	}

	// Check user-provided services with checkmarx tags
	if userProvided, ok := services["user-provided"]; ok {
		for _, service := range userProvided {
			if tags, ok := service["tags"].([]interface{}); ok {
				for _, tag := range tags {
					if tagStr, ok := tag.(string); ok {
						if strings.Contains(strings.ToLower(tagStr), "checkmarx") {
							return true
						}
					}
				}
			}
		}
	}

	return false
}

// CheckmarxCredentials holds Checkmarx IAST credentials
type CheckmarxCredentials struct {
	URL        string // Agent download URL
	ManagerURL string // Checkmarx manager URL
	APIKey     string // API key for authentication
}

// getCredentials retrieves Checkmarx IAST credentials from service binding
func (c *CheckmarxIASTAgentFramework) getCredentials() CheckmarxCredentials {
	creds := CheckmarxCredentials{}

	vcapServices := os.Getenv("VCAP_SERVICES")
	if vcapServices == "" {
		return creds
	}

	var services map[string][]map[string]interface{}
	if err := json.Unmarshal([]byte(vcapServices), &services); err != nil {
		return creds
	}

	// Look for checkmarx-iast service
	serviceNames := []string{
		"checkmarx-iast",
		"checkmarx",
		"user-provided",
	}

	for _, serviceName := range serviceNames {
		if serviceList, ok := services[serviceName]; ok {
			for _, service := range serviceList {
				if credentials, ok := service["credentials"].(map[string]interface{}); ok {
					// Get agent download URL
					if url, ok := credentials["url"].(string); ok {
						creds.URL = url
					} else if url, ok := credentials["agent_url"].(string); ok {
						creds.URL = url
					}

					// Get manager URL
					if managerURL, ok := credentials["manager_url"].(string); ok {
						creds.ManagerURL = managerURL
					} else if managerURL, ok := credentials["managerUrl"].(string); ok {
						creds.ManagerURL = managerURL
					}

					// Get API key
					if apiKey, ok := credentials["api_key"].(string); ok {
						creds.APIKey = apiKey
					} else if apiKey, ok := credentials["apiKey"].(string); ok {
						creds.APIKey = apiKey
					}

					if creds.URL != "" {
						return creds
					}
				}
			}
		}
	}

	return creds
}

// downloadAgent downloads the agent JAR from the given URL
func (c *CheckmarxIASTAgentFramework) downloadAgent(url, destPath string) error {
	c.context.Log.Debug("Downloading Checkmarx IAST agent from %s", url)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download agent: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download agent: HTTP %d", resp.StatusCode)
	}

	outFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, resp.Body); err != nil {
		return fmt.Errorf("failed to write agent file: %w", err)
	}

	return nil
}

// appendToJavaOpts appends a value to JAVA_OPTS
func (c *CheckmarxIASTAgentFramework) appendToJavaOpts(value string) error {
	javaOptsFile := filepath.Join(c.context.Stager.DepDir(), "env", "JAVA_OPTS")

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
	return c.context.Stager.WriteEnvFile(javaOptsFile, existingOpts)
}
