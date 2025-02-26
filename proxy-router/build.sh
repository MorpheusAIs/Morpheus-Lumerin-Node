#!/bin/sh
echo "Inbound Tag: $TAG_NAME" 
echo "Inbound Commit: $COMMIT" 
# Check if TAG_NAME is set; if not, use the latest Git tag or fallback to 0.1.0
if [ -z "$TAG_NAME" ]; then
  TAG_NAME=$(git describe --tags --abbrev=0 2>/dev/null || echo "0.1.0")
  if [ "$TAG_NAME" = "0.1.0" ]; then
    echo "Warning: No Git tags found. Defaulting to TAG_NAME=$TAG_NAME"
  else
    echo "Using latest Git tag: $TAG_NAME"
  fi
fi

VERSION=$TAG_NAME
echo VERSION=$VERSION
# if commit is not set, use the latest commit
if [ -z "$COMMIT" ]; then
  COMMIT=$(git rev-parse HEAD)
fi
echo COMMIT=$COMMIT
go mod tidy 
go build \
  -tags docker \
  -ldflags="-s -w \
    -X 'github.com/Lumerin-protocol/Morpheus-Lumerin-Node/proxy-router/internal/config.BuildVersion=$VERSION' \
    -X 'github.com/Lumerin-protocol/Morpheus-Lumerin-Node/proxy-router/internal/config.Commit=$COMMIT' \
    -X 'github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config.BuildVersion=$VERSION' \
    -X 'github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config.Commit=$COMMIT' \
  " \
  -o ./proxy-router cmd/main.go
