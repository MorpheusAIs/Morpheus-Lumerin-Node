run:
	GOTRACEBACK=crash go run -gcflags '-N -l' -race cmd/main.go

build:
	./build.sh
	
clean:
	rm -rf bin logs

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