#!/usr/bin/env bash

set -euo pipefail

cd java-buildpack

echo "Installing dependencies..."
bundle install

echo "Building online buildpack for release..."
OFFLINE=false bundle exec rake package

# Get version
VERSION=$(cat ../buildpack-version/version)

# Prepare release artifacts
cp build/java-buildpack-*.zip ../release/java-buildpack-v${VERSION}.zip

# Create release metadata
echo "v${VERSION}" > ../release/tag
echo "Java Buildpack v${VERSION}" > ../release/name

# Create release notes
cat > ../release/body <<EOF
## Java Buildpack v${VERSION}

This release includes the latest dependencies and improvements.

### Installation

Download the buildpack and install it to your Cloud Foundry:

\`\`\`bash
cf create-buildpack java_buildpack java-buildpack-v${VERSION}.zip 1 --enable
\`\`\`

### Buildpack Artifacts

- **java-buildpack-v${VERSION}.zip** - Online buildpack (downloads dependencies at staging time)

### Documentation

For full documentation, see [cloudfoundry.org/java-buildpack](https://github.com/cloudfoundry/java-buildpack)
EOF

echo "Release artifacts prepared for version ${VERSION}"
ls -lh ../release/
