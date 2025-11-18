# Integration Tests

This directory contains integration tests for the Java buildpack using the [Switchblade](https://github.com/cloudfoundry/switchblade) framework.

## Overview

Switchblade is a Go-based integration testing framework that supports both Cloud Foundry and Docker platforms. This allows us to write tests once and run them on either platform.

## Prerequisites

- Go 1.25 or later
- Cloud Foundry CLI (if testing on CF)
- Docker (if testing on Docker)
- A packaged buildpack zip file
- **GitHub Personal Access Token** (required for Docker platform tests)
  - Create token at: https://github.com/settings/tokens
  - Requires `public_repo` or `repo` scope
  - Used to query buildpack metadata from GitHub API

## Running Tests

### Package the Buildpack

First, create a buildpack zip file:

```bash
bundle exec rake package
```

This will create a file like `java-buildpack-v4.x.x.zip` in the project root.

### Run Integration Tests

Use the provided script to run the tests:

```bash
# Test on Cloud Foundry (default)
BUILDPACK_FILE=/path/to/java-buildpack-v4.x.x.zip ./scripts/integration.sh

# Test on Docker (requires GitHub token)
BUILDPACK_FILE=/path/to/java-buildpack-v4.x.x.zip \
GITHUB_TOKEN=your_github_token_here \
./scripts/integration.sh --platform docker

# Run cached/offline tests
BUILDPACK_FILE=/path/to/java-buildpack-v4.x.x.zip ./scripts/integration.sh --cached

# Specify a different stack
BUILDPACK_FILE=/path/to/java-buildpack-v4.x.x.zip ./scripts/integration.sh --stack cflinuxfs4
```

### Run Tests Directly with Go

You can also run the tests directly using Go:

```bash
cd src/integration

# Run all tests
BUILDPACK_FILE=/path/to/buildpack.zip go test -v -timeout 30m

# Run specific test suite
BUILDPACK_FILE=/path/to/buildpack.zip go test -v -run TestIntegration/Tomcat

# Run on Docker
BUILDPACK_FILE=/path/to/buildpack.zip go test -v -platform=docker

# Run offline tests
BUILDPACK_FILE=/path/to/buildpack.zip go test -v -cached
```

## Test Organization

### Test Files

- `init_test.go` - Test suite initialization and configuration
- `tomcat_test.go` - Tomcat container tests
- `spring_boot_test.go` - Spring Boot application tests
- `java_main_test.go` - Java Main class application tests
- `offline_test.go` - Offline/cached buildpack tests

### Test Fixtures

Tests use fixtures from the `spec/fixtures` directory. The main fixture for integration tests is:
- `integration_valid` - A simple Java application with a Main-Class

## Configuration

### Environment Variables

- `BUILDPACK_FILE` (required) - Path to the packaged buildpack zip file
- `PLATFORM` - Platform to test against: `cf` (default) or `docker`
- `STACK` - Stack to use for tests (default: `cflinuxfs4`)
- `CACHED` - Run offline/cached tests (default: `false`)
- `GITHUB_TOKEN` - GitHub API token to avoid rate limiting

### Command-Line Flags

- `-platform` - Platform type (`cf` or `docker`)
- `-stack` - Stack name (e.g., `cflinuxfs4`)
- `-cached` - Enable offline tests
- `-github-token` - GitHub API token
- `-serial` - Run tests serially instead of in parallel

## Test Coverage

The integration tests cover:

1. **Container Types**
   - Tomcat container with WAR files
   - Spring Boot executable JARs
   - Java Main applications

2. **JRE Selection**
   - Java 8, 11, 17 runtime selection
   - Multiple JRE vendors (OpenJDK, Zulu, etc.)

3. **Configuration**
   - Memory calculator settings
   - Custom JAVA_OPTS
   - Framework-specific configuration

4. **Offline Mode**
   - Cached buildpack deployment
   - No internet access scenarios

## Writing New Tests

To add a new test:

1. Create a new test file in `src/integration/` (e.g., `myfeature_test.go`)
2. Define a test function that returns `func(*testing.T, spec.G, spec.S)`:
   ```go
   func testMyFeature(platform switchblade.Platform, fixtures string) func(*testing.T, spec.G, spec.S) {
       return func(t *testing.T, context spec.G, it spec.S) {
           // Your tests here
       }
   }
   ```
3. Register the test in `init_test.go`:
   ```go
   suite("MyFeature", testMyFeature(platform, fixtures))
   ```

## CI/CD Integration

To integrate with CI/CD pipelines:

```yaml
# Example GitHub Actions
- name: Run Integration Tests
  env:
    BUILDPACK_FILE: ${{ github.workspace }}/java-buildpack.zip
    CF_API: ${{ secrets.CF_API }}
    CF_USERNAME: ${{ secrets.CF_USERNAME }}
    CF_PASSWORD: ${{ secrets.CF_PASSWORD }}
  run: |
    ./scripts/integration.sh --platform cf
```

## Comparison to Old Tests

The previous integration tests were:
- Located in a separate repository (`java-buildpack-system-test`)
- Written in Java with JUnit
- Only supported Cloud Foundry
- Required extensive configuration

The new Switchblade-based tests:
- Are co-located with the buildpack code
- Written in Go with Gomega matchers
- Support both Cloud Foundry and Docker
- Have simpler configuration and setup

## Troubleshooting

### Tests fail to compile
```bash
go mod tidy
go mod download
```

### Buildpack not found
Ensure the `BUILDPACK_FILE` environment variable points to a valid zip file:
```bash
ls -lh $BUILDPACK_FILE
```

### CF login issues
Ensure you're logged into Cloud Foundry:
```bash
cf login -a <api-endpoint>
```

### Docker issues
Ensure Docker is running and you have permission to use it:
```bash
docker ps
```

### GitHub authentication errors with Docker platform
If you see errors like "Bad credentials" or "401 Unauthorized" when running Docker platform tests:

```
failed to build buildpacks: failed to list buildpacks: received unexpected response status: HTTP/2.0 401 Unauthorized
```

This means you need to provide a GitHub Personal Access Token:

1. Create a token at https://github.com/settings/tokens
2. Grant it `public_repo` or `repo` scope
3. Export it as an environment variable:
   ```bash
   export GITHUB_TOKEN=your_token_here
   BUILDPACK_FILE=/path/to/buildpack.zip ./scripts/integration.sh --platform docker
   ```

Alternatively, pass it via the command line:
```bash
BUILDPACK_FILE=/path/to/buildpack.zip ./scripts/integration.sh --platform docker --github-token your_token_here
```

## References

- [Switchblade Documentation](https://github.com/cloudfoundry/switchblade)
- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [Gomega Matchers](https://onsi.github.io/gomega/)
- [Cloud Foundry Buildpack Documentation](https://docs.cloudfoundry.org/buildpacks/)
