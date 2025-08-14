BINARY_NAME=nativefire
VERSION=1.0.0
BUILD_DIR=build

.PHONY: build clean test install

build:
	mkdir -p $(BUILD_DIR)
	go build -ldflags="-X github.com/clix-so/nativefire/cmd.Version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME) .

build-all:
	mkdir -p $(BUILD_DIR)
	# Linux
	GOOS=linux GOARCH=amd64 go build -ldflags="-X github.com/clix-so/nativefire/cmd.Version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build -ldflags="-X github.com/clix-so/nativefire/cmd.Version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 .
	# macOS
	GOOS=darwin GOARCH=amd64 go build -ldflags="-X github.com/clix-so/nativefire/cmd.Version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -ldflags="-X github.com/clix-so/nativefire/cmd.Version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 .
	# Windows
	GOOS=windows GOARCH=amd64 go build -ldflags="-X github.com/clix-so/nativefire/cmd.Version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe .
	GOOS=windows GOARCH=arm64 go build -ldflags="-X github.com/clix-so/nativefire/cmd.Version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-windows-arm64.exe .

test:
	go test -v ./...

clean:
	rm -rf $(BUILD_DIR)

install: build
	cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/

uninstall:
	rm -f /usr/local/bin/$(BINARY_NAME)