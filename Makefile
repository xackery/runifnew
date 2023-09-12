NAME ?= runifnew
VERSION := 0.1.6

sanitize:
	rm -rf vendor/
	go vet -tags ci ./...
	test -z $(goimports -e -d . | tee /dev/stderr)
	gocyclo -over 30 .
	golint -set_exit_status $(go list -tags ci ./...)
	staticcheck -go 1.14 ./...
	go test -tags ci -covermode=atomic -coverprofile=coverage.out ./...
    coverage=`go tool cover -func coverage.out | grep total | tr -s '\t' | cut -f 3 | grep -o '[^%]*'`
run: build-darwin
	cd bin && mkdir -p rof
	cd bin/rof && echo "[LoginServer]" > eqhost.txt && echo "host=test.com:9000" >> eqhost.txt
	cd bin && ./filelistbuilder-darwin-x64 rof https://test.com filelistbuilder-darwin-x64

.PHONY: build-all
build-all: sanitize build-prepare build-linux build-darwin build-windows	
.PHONY: build-prepare
build-prepare:
	@echo "Preparing talkeq ${VERSION}"
	@rm -rf bin/*
	@-mkdir -p bin/
.PHONY: build-darwin
build-darwin:
	@echo "Building darwin ${VERSION}"
	@GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -buildmode=pie -ldflags="-X main.Version=${VERSION} -s -w" -o bin/${NAME}-darwin main.go
.PHONY: build-linux
build-linux:
	@echo "Building Linux ${VERSION}"
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -buildmode=pie -ldflags="-X main.Version=${VERSION} -w" -o bin/${NAME}-linux main.go		
.PHONY: build-windows
build-windows:
	@echo "Building Windows ${VERSION}"
	@GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -buildmode=pie -ldflags="-X main.Version=${VERSION} -s -w" -o bin/${NAME}.exe main.go


# CICD triggers this
.PHONY: set-variable
set-version:
	@echo "VERSION=${VERSION}" >> $$GITHUB_ENV
	