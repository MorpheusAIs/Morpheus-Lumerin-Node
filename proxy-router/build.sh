#!/bin/sh
# Check if TAG_NAME is set; if not, look up the existing tag from the repository
# This is READ-ONLY: finds tags created by CI/CD, does NOT calculate new versions
if [ -z "$TAG_NAME" ]; then
  # Get current branch name
  CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
  
  # Try to find exact tag for current commit first
  TAG_NAME=$(git describe --exact-match --tags HEAD 2>/dev/null)
  
  if [ -n "$TAG_NAME" ]; then
    echo "✅ Found exact tag for current commit: $TAG_NAME"
  else
    # No exact tag, find the most recent tag for this branch
    case "$CURRENT_BRANCH" in
      main)
        # MAIN branch: Find latest production tag (no suffix, format: v*.*.*)
        # Looks for tags on this branch that match the pattern v[1-9]*.*.*
        TAG_NAME=$(git describe --tags --abbrev=0 --match='v[1-9]*' HEAD 2>/dev/null | grep -v '\-' | head -1)
        if [ -n "$TAG_NAME" ]; then
          echo "✅ MAIN branch - Using latest production tag: $TAG_NAME"
        else
          TAG_NAME="v0.1.0"
          echo "⚠️  MAIN branch - No production tags found, defaulting to: $TAG_NAME"
        fi
        ;;
      test)
        # TEST branch: Find latest test tag (format: v*.*.*-test)
        TAG_NAME=$(git describe --tags --abbrev=0 --match='v[1-9]*-test' HEAD 2>/dev/null)
        if [ -n "$TAG_NAME" ]; then
          echo "✅ TEST branch - Using latest test tag: $TAG_NAME"
        else
          TAG_NAME="v0.1.0-test"
          echo "⚠️  TEST branch - No test tags found, defaulting to: $TAG_NAME"
        fi
        ;;
      *)
        # Other branches: Use commit hash as version (no guessing at tags)
        COMMIT_HASH=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
        TAG_NAME="dev-${COMMIT_HASH}"
        echo "ℹ️  Branch '$CURRENT_BRANCH' - Using commit-based version: $TAG_NAME"
        ;;
    esac
  fi
fi

VERSION=$TAG_NAME
echo VERSION=$VERSION
# if commit is not set, use the latest commit
if [ -z "$COMMIT" ]; then
  COMMIT=$(git rev-parse HEAD)
fi
echo COMMIT=$COMMIT
# go mod tidy already handled in Dockerfile via go mod download
go build \
  -tags docker \
  -ldflags="-s -w \
    -X 'github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config.BuildVersion=$VERSION' \
    -X 'github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config.Commit=$COMMIT' \
  " \
  -o ./proxy-router cmd/main.go
