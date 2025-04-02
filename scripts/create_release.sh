#!/bin/bash
# scripts/create_release.sh - Helper script for creating a new GoCard release

# Configuration
VERSION="0.3.0"
TAG_NAME="v$VERSION"
BRANCH_NAME="release/$VERSION"

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check if we're on a clean branch
if [ -n "$(git status --porcelain)" ]; then
    echo -e "${RED}Error: Working directory is not clean.${NC}"
    echo "Please commit or stash all changes before creating a release."
    exit 1
fi

# Create a release branch
echo -e "${GREEN}Creating release branch $BRANCH_NAME...${NC}"
git checkout -b "$BRANCH_NAME"

# Update version in files
echo -e "${GREEN}Updating version to $VERSION in files...${NC}"
sed -i 's/Version = "0.1.0"/Version = "'"$VERSION"'"/' cmd/gocard/flags.go

# Confirm CHANGELOG.md exists
if [ ! -f CHANGELOG.md ]; then
    echo -e "${YELLOW}Warning: CHANGELOG.md not found.${NC}"
    echo "Consider creating a CHANGELOG.md file to document changes."
    read -p "Continue without CHANGELOG? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        git checkout -
        exit 1
    fi
fi

# Check for RELEASE_NOTES.md
if [ ! -f RELEASE_NOTES.md ]; then
    echo -e "${YELLOW}Warning: RELEASE_NOTES.md not found.${NC}"
    echo "Consider creating release notes to document this version."
    read -p "Continue without RELEASE_NOTES? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        git checkout -
        exit 1
    fi
fi

# Commit the version changes
echo -e "${GREEN}Committing version changes...${NC}"
git add cmd/gocard/flags.go CHANGELOG.md RELEASE_NOTES.md
git commit -m "chore: bump version to $VERSION for release"

# Create and push the tag
echo -e "${GREEN}Creating and pushing tag $TAG_NAME...${NC}"
git tag -a "$TAG_NAME" -m "Release $VERSION"
git push origin "$TAG_NAME"

# Push the release branch
echo -e "${GREEN}Pushing release branch $BRANCH_NAME...${NC}"
git push -u origin "$BRANCH_NAME"

echo -e "${GREEN}Release process initiated!${NC}"
echo "The release workflow should now be running on GitHub."
echo "You can check the progress at: https://github.com/DavidMiserak/GoCard/actions"
echo -e "${YELLOW}Note: You may want to create a PR to merge these changes back to main.${NC}"
