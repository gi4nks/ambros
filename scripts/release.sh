#!/bin/bash
set -e

# release.sh - Automated release script for Ambros
# Usage: ./release.sh v3.0.3 "Release message"

if [ $# -lt 1 ]; then
    echo "Usage: $0 <version> [release_message]"
    echo "Example: $0 v3.0.3 'Bug fixes and improvements'"
    exit 1
fi

VERSION=$1
RELEASE_MESSAGE=${2:-"Release $VERSION"}

# Remove 'v' prefix for version.txt if present
VERSION_NUMBER=${VERSION#v}

echo "🚀 Releasing Ambros $VERSION"

# Update embedded version
echo "$VERSION" > cmd/commands/version.txt
echo "✅ Updated version.txt to $VERSION"

# Run tests
echo "🧪 Running tests..."
make test

# Build and verify
echo "🔨 Building and verifying..."
make build
./bin/ambros version

# Commit changes
echo "📝 Committing changes..."
git add cmd/commands/version.txt
git commit -m "bump: Update to $VERSION"

# Create tag
echo "🏷️  Creating tag $VERSION..."
git tag "$VERSION" -m "$RELEASE_MESSAGE"

# Push changes and tags
echo "📤 Pushing to origin..."
git push origin master --tags

echo "✅ Release $VERSION complete!"
echo ""
echo "🔗 Install command for users:"
echo "go install github.com/gi4nks/ambros/v3@$VERSION"
echo ""
echo "🔗 Latest install command:"
echo "go install github.com/gi4nks/ambros/v3@latest"