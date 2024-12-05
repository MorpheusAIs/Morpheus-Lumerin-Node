#!/bin/sh

# This script assumes you've cloned the repository and want to build with proper commit and version numbers
# It will use the latest Git tag as the version number and the latest short commit hash as the commit number
# It also leverages the docker-compose.yml file to build the Docker image

# Pre-requisites: Docker & Git installed

# Assumptions: 
# - properly formatted .env in current directory 
# - properly formatted models-config.json in current directory 
# - properly formatted rating-config.json in current directory

# Check if TAG_NAME is set; if not, use the latest Git tag
if [ -z "$TAG_NAME" ]; then
  VLAST=$(git describe --tags --abbrev=0 --match='v[1-9]*' refs/remotes/origin/main 2>/dev/null | cut -c2-)
  [ $VLAST ] && declare $(echo $VLAST | awk -F '.' '{print "VMAJ="$1" VMIN="$2" VPAT="$3}')
  MB=$(git merge-base refs/remotes/origin/main HEAD)
  VPAT=$(git rev-list --count --no-merges ${MB}..HEAD)
  TAG_NAME=${VMAJ}.${VMIN}.${VPAT}
fi
VERSION=$TAG_NAME
echo VERSION=$VERSION

# if commit is not set, use the latest commit
if [ -z "$COMMIT" ]; then
  SHORT_COMMIT=$(git rev-parse --short HEAD)
fi
COMMIT=$SHORT_COMMIT
echo COMMIT=$COMMIT
export VERSION COMMIT TAG_NAME

# Check if the user wants to build or run the Docker image
if [ "$1" = "--build" ]; then
  echo "Building Docker image..."
  docker-compose build
  docker tag proxy-router:$VERSION proxy-router:latest
elif [ "$1" = "--run" ]; then
  echo "Running Docker container..."
  docker-compose up
else
  echo "Usage: $0 [--build | --run]"
fi