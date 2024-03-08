VERSION=$(grep '^VERSION=' .version | cut -d '=' -f 2-)
echo VERSION=$VERSION

# if commit is not set, use the latest commit
if [ -z "$COMMIT" ]; then
  COMMIT=$(git rev-parse HEAD)
fi
echo COMMIT=$COMMIT

go build \
  -ldflags="-s -w \
    -X 'gitlab.com/TitanInd/proxy/proxy-router-v3/internal/config.BuildVersion=$VERSION' \
    -X 'gitlab.com/TitanInd/proxy/proxy-router-v3/internal/config.Commit=$COMMIT' \
  " \
  -o bin/hashrouter cmd/main.go 
