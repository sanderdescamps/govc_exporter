#!/usr/bin/bash

# Fetch latest version
TAG=$(curl -s https://api.github.com/repos/prometheus/prometheus/releases | jq -r '.[] | .tag_name' | head -n 1)
echo $TAG

# Remove the `v` prefix
VERSION=$(echo $TAG | sed 's/v//')
echo $VERSION

# Check CPU architecture
ARCH=$(uname -s | tr A-Z a-z)-$(uname -m | sed 's/x86_64/amd64/')
echo $ARCH

# Download the release
curl -L -O https://github.com/prometheus/prometheus/releases/download/$TAG/prometheus-$VERSION.$ARCH.tar.gz

# Extract the file
tar -xvzf prometheus-$VERSION.$ARCH.tar.gz

# Move promtool
mv prometheus-$VERSION.$ARCH/promtool ~/.local/bin

# Verify if promtool is installed
promtool --version

# Delete downloaded files
rm -rf prometheus-$VERSION.$ARCH
rm -rf prometheus-$VERSION.$ARCH.tar.gz