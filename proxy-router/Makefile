run:
	go run cmd/main.go

run-user:
	WALLET_PRIVATE_KEY=0x5de4111afa1a4b94908f83103eb1f1706367c2e68ca870fc3fb9a804cdab365a DIAMOND_CONTRACT_ADDRESS=0x6aBE1d282f72B474E54527D93b979A4f64d3030a MOR_TOKEN_ADDRESS=0x7431aDa8a591C955a994a21710752EF9b882b8e3 PROXY_STORAGE_PATH='./data/badger2/' PROXY_ADDRESS='0.0.0.0:3334' WEB_ADDRESS='0.0.0.0:8083' BLOCKSCOUT_API_URL=https://base.blockscout.com/api/v2 ETH_NODE_CHAIN_ID=8453 make run

run-race:
	GOTRACEBACK=crash go run -gcflags '-N -l' -race cmd/main.go

build:
	TAG_NAME=$(TAG_NAME) ./build.sh
	cp .env.example .env 
	cp models-config.json.example models-config.json
	cp rating-config.json.example rating-config.json
	
clean:
	rm -rf bin logs data .cookie proxy.conf proxy-router .env models-config.json rating-config.json

install:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.53.3
	go install golang.org/x/vuln/cmd/govulncheck@latest
	go install github.com/praetorian-inc/gokart@latest
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	
format:
	gofmt -s -w .
	
lint:
	golangci-lint run
	govulncheck ./...
	gokart scan .
	gosec ./...

test-unit:
	go test -v -p 1 $$(go list ./... | grep -v /test)

test-integration:
	go test -v ./test/...

swagger:
	swag fmt -g ./internal/handlers/httphandlers/http.go && swag init -g ./internal/handlers/httphandlers/http.go --parseInternal --parseDependency

mocks:
	mockery

reset-keychain:
	go run cmd/main.go -app-reset-keychain

update-bindings:
	cd ../smart-contracts && yarn bindings-go && cp -r bindings/go/contracts/{diamond/facets/*,tokens/*,mock/tokens/*} ../proxy-router/internal/repositories/contracts/bindings
