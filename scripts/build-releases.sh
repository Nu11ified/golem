#!/bin/bash

# Build script for creating distributable golem binaries
# This script builds the golem CLI for multiple platforms

set -e

# Get version from the CLI
VERSION="v0.1.0"
BINARY_NAME="golem"
OUTPUT_DIR="releases"

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Build configurations: OS/ARCH
declare -a PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
)

echo "🔨 Building $BINARY_NAME $VERSION for multiple platforms..."

for platform in "${PLATFORMS[@]}"; do
    IFS='/' read -r -a platform_split <<< "$platform"
    GOOS="${platform_split[0]}"
    GOARCH="${platform_split[1]}"
    
    # Set output filename
    output_name="$BINARY_NAME-$VERSION-$GOOS-$GOARCH"
    if [ "$GOOS" = "windows" ]; then
        output_name="$output_name.exe"
    fi
    
    echo "Building for $GOOS/$GOARCH..."
    
    # Build binary
    env GOOS="$GOOS" GOARCH="$GOARCH" go build -o "$OUTPUT_DIR/$output_name" -ldflags="-s -w" ./cmd/golem/main.go
    
    # Create compressed archive
    if [ "$GOOS" = "windows" ]; then
        (cd "$OUTPUT_DIR" && zip "$BINARY_NAME-$VERSION-$GOOS-$GOARCH.zip" "$output_name")
    else
        (cd "$OUTPUT_DIR" && tar -czf "$BINARY_NAME-$VERSION-$GOOS-$GOARCH.tar.gz" "$output_name")
    fi
    
    echo "✅ Built $output_name"
done

echo ""
echo "🎉 All binaries built successfully!"
echo "📦 Files created in $OUTPUT_DIR:"
ls -la "$OUTPUT_DIR"

echo ""
echo "📋 To test locally:"
echo "   chmod +x $OUTPUT_DIR/golem-$VERSION-$(go env GOOS)-$(go env GOARCH)"
echo "   $OUTPUT_DIR/golem-$VERSION-$(go env GOOS)-$(go env GOARCH) version" 