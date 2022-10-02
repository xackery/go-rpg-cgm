NAME := cgm
VERSION ?= 0.1.0

run:
	cd cmd && go run main.go

build: sanitize build-prepare build-darwin

.PHONY: build-prepare
build-prepare:
	@echo "Preparing ${NAME} ${VERSION}"
	@rm -rf bin/*
	@-mkdir -p bin/

.PHONY: build-darwin
build-darwin:
	@#echo "Building darwin ${VERSION}"
	@GOOS=darwin GOARCH=amd64 go build -buildmode=pie -ldflags="-X main.Version=${VERSION} -s -w" -o bin/${NAME}-${VERSION}-darwin-x64 cmd/main.go

.PHONY: build-linux
build-linux:
	@echo "Building Linux ${VERSION}"
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -buildmode=pie -ldflags="-X main.Version=${VERSION} -w" -o bin/${NAME}-${VERSION}-linux-x64 cmd/main.go		

.PHONY: build-windows
build-windows:
	@echo "Building Windows ${VERSION}"
	@GOOS=windows GOARCH=386 CGO_ENABLED=1 CC=i686-w64-mingw32-gcc go build -buildmode=pie -ldflags="-X main.Version=${VERSION} -s -w" -o bin/${NAME}-${VERSION}-win-x86.exe cmd/main.go
	@GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc go build -buildmode=pie -ldflags="-X main.Version=${VERSION} -s -w" -o bin/${NAME}-${VERSION}-win-x64.exe cmd/main.go

.PHONY: sanitize
sanitize:
	@goimports -w .
	@golint		