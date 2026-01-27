#!/bin/bash
# Version Management Script for Pryx
# Supports: major, minor, patch version bumping

set -e

VERSION_FILE="VERSION"
CHANGELOG_FILE="CHANGELOG.md"

# Read current version
if [ ! -f "$VERSION_FILE" ]; then
    echo "0.1.0" > "$VERSION_FILE"
    echo "Created VERSION file with initial version 0.1.0"
fi

CURRENT_VERSION=$(cat "$VERSION_FILE")
echo "Current version: $CURRENT_VERSION"

# Parse version into components
IFS='.' read -r MAJOR MINOR PATCH <<< "$CURRENT_VERSION"

# Bump version based on argument
case "$1" in
    major)
        MAJOR=$((MAJOR + 1))
        MINOR=0
        PATCH=0
        ;;
    minor)
        MINOR=$((MINOR + 1))
        PATCH=0
        ;;
    patch)
        PATCH=$((PATCH + 1))
        ;;
    *)
        echo "Usage: $0 {major|minor|patch}"
        exit 1
        ;;
esac

# Create new version
NEW_VERSION="${MAJOR}.${MINOR}.${PATCH}"
echo "New version: $NEW_VERSION"

# Write new version to file
echo "$NEW_VERSION" > "$VERSION_FILE"
echo "✓ Updated VERSION to $NEW_VERSION"

# Add entry to CHANGELOG.md (if exists)
if [ -f "$CHANGELOG_FILE" ]; then
    DATE=$(date +%Y-%m-%d)
    NEW_ENTRY="## [$NEW_VERSION] - $DATE

### Changes
- [TODO: Describe changes]

"

    # Insert after the first line if CHANGELOG exists
    if [ -s "$CHANGELOG_FILE" ]; then
        TMP_FILE=$(mktemp)
        echo "$NEW_ENTRY" > "$TMP_FILE"
        cat "$CHANGELOG_FILE" >> "$TMP_FILE"
        mv "$TMP_FILE" "$CHANGELOG_FILE"
        echo "✓ Updated CHANGELOG.md"
    fi
else
    # Create new CHANGELOG.md
    echo "# Changelog

All notable changes to Pryx will be documented in this file.

$NEW_ENTRY" > "$CHANGELOG_FILE"
    echo "✓ Created CHANGELOG.md with new entry"
fi

echo ""
echo "Next steps:"
echo "  1. Review CHANGELOG.md and add actual changes"
echo "  2. Run 'git add VERSION CHANGELOG.md'"
echo "  3. Run 'git commit -m \"chore: bump version to $NEW_VERSION\"'"
echo "  4. Run 'make version-tag'"
echo "  5. Run 'make version-push'"
