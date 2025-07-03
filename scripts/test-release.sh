#!/bin/bash

# Test release script - validates that everything is ready for automated release
# This script helps ensure your commit will trigger a successful release

set -e

echo "🔍 Testing release process locally..."

# Check if we're in the right directory
if [ ! -f "go.mod" ]; then
    echo "❌ Error: Please run this script from the project root directory"
    exit 1
fi

# Check Go version
echo "📋 Checking Go version..."
go version

# Check if git is clean
echo "🔍 Checking git status..."
if [ -n "$(git status --porcelain)" ]; then
    echo "⚠️  Warning: You have uncommitted changes"
    git status --short
else
    echo "✅ Git working directory is clean"
fi

# Get current version info
echo "📊 Version information:"
LATEST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "No tags found")
echo "   Latest tag: $LATEST_TAG"

COMMIT_COUNT=$(git rev-list --count HEAD 2>/dev/null || echo "0")
echo "   Total commits: $COMMIT_COUNT"

if [ "$LATEST_TAG" != "No tags found" ]; then
    COMMITS_SINCE_TAG=$(git rev-list --count ${LATEST_TAG}..HEAD 2>/dev/null || echo "0")
    echo "   Commits since last tag: $COMMITS_SINCE_TAG"
fi

# Test dependencies
echo "🔧 Testing dependencies..."
go mod download
go mod tidy
go mod verify
echo "✅ Dependencies are valid"

# Run tests
echo "🧪 Running tests..."
if go test -v ./...; then
    echo "✅ All tests passed"
else
    echo "❌ Tests failed - fix before releasing"
    exit 1
fi

# Test build
echo "🏗️  Testing build process..."
if go build -v ./cmd/golem; then
    echo "✅ Build successful"
    rm -f golem
else
    echo "❌ Build failed - fix before releasing"
    exit 1
fi

# Test cross-compilation
echo "🌐 Testing cross-compilation..."
PLATFORMS="linux/amd64 darwin/amd64 windows/amd64"
for platform in $PLATFORMS; do
    IFS='/' read -r -a platform_split <<< "$platform"
    GOOS="${platform_split[0]}"
    GOARCH="${platform_split[1]}"
    
    echo "   Testing $GOOS/$GOARCH..."
    if env GOOS="$GOOS" GOARCH="$GOARCH" go build -o "test-$GOOS-$GOARCH" ./cmd/golem; then
        rm -f "test-$GOOS-$GOARCH"*
        echo "   ✅ $GOOS/$GOARCH build successful"
    else
        echo "   ❌ $GOOS/$GOARCH build failed"
        exit 1
    fi
done

# Check last commit message
echo "💬 Checking commit message for release triggers..."
LAST_COMMIT=$(git log -1 --pretty=format:"%s" 2>/dev/null || echo "No commits")
echo "   Last commit: $LAST_COMMIT"

# Analyze commit message for release patterns
WILL_RELEASE=false
RELEASE_TYPE="none"

if [[ "$LAST_COMMIT" =~ ^(release|Release|RELEASE)[:\ ] ]] || \
   [[ "$LAST_COMMIT" =~ \[release\] ]] || \
   [[ "$LAST_COMMIT" =~ ^(v[0-9]+\.[0-9]+\.[0-9]+) ]]; then
    WILL_RELEASE=true
    RELEASE_TYPE="explicit"
elif [[ "$LAST_COMMIT" =~ ^(feat|fix|breaking)(\(.+\))?!: ]]; then
    WILL_RELEASE=true
    if [[ "$LAST_COMMIT" =~ ^breaking[:\ ] ]] || [[ "$LAST_COMMIT" =~ !: ]]; then
        RELEASE_TYPE="major"
    elif [[ "$LAST_COMMIT" =~ ^feat[:\ ] ]]; then
        RELEASE_TYPE="minor"
    else
        RELEASE_TYPE="patch"
    fi
elif [[ "$LAST_COMMIT" =~ ^(feat|feature)[:\ ] ]]; then
    WILL_RELEASE=true
    RELEASE_TYPE="minor"
elif [[ "$LAST_COMMIT" =~ ^fix[:\ ] ]]; then
    WILL_RELEASE=true
    RELEASE_TYPE="patch"
fi

if [ "$WILL_RELEASE" = true ]; then
    echo "🚀 This commit WILL trigger a release!"
    echo "   Release type: $RELEASE_TYPE"
    
    # Calculate what the new version would be
    if [ "$LATEST_TAG" != "No tags found" ]; then
        BASE_VERSION=${LATEST_TAG#v}
        IFS='.' read -r -a VERSION_PARTS <<< "$BASE_VERSION"
        MAJOR=${VERSION_PARTS[0]}
        MINOR=${VERSION_PARTS[1]}
        PATCH=${VERSION_PARTS[2]}
        
        case "$RELEASE_TYPE" in
            "major")
                MAJOR=$((MAJOR + 1))
                MINOR=0
                PATCH=0
                ;;
            "minor")
                MINOR=$((MINOR + 1))
                PATCH=0
                ;;
            "patch")
                PATCH=$((PATCH + 1))
                ;;
        esac
        
        NEW_VERSION="v${MAJOR}.${MINOR}.${PATCH}"
        echo "   Expected new version: $NEW_VERSION"
    fi
else
    echo "ℹ️  This commit will NOT trigger a release"
    echo "   Use commit patterns like:"
    echo "   - 'feat: add new feature' (minor release)"
    echo "   - 'fix: resolve bug' (patch release)"
    echo "   - 'feat!: breaking change' (major release)"
    echo "   - 'release: v1.2.3' (explicit release)"
fi

echo ""
echo "🎉 Release test completed successfully!"
echo ""
echo "📋 Summary:"
echo "   Tests: ✅ Passed"
echo "   Build: ✅ Successful"
echo "   Cross-compilation: ✅ Working"
echo "   Dependencies: ✅ Valid"
echo "   Release trigger: $([ "$WILL_RELEASE" = true ] && echo "✅ Yes ($RELEASE_TYPE)" || echo "❌ No")"
echo ""

if [ "$WILL_RELEASE" = true ]; then
    echo "🚀 Ready to push! This will trigger a release."
else
    echo "💡 Ready to push! (No release will be triggered)"
fi 