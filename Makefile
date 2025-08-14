BINARY_NAME=nativefire
VERSION=1.0.0
BUILD_DIR=build

.PHONY: build clean test install lint lint-fix format check

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

# Linting and formatting
format:
	@echo "🎨 Formatting code..."
	gofmt -w .
	@echo "✅ Code formatted!"

lint:
	@echo "🔍 Running lint checks..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --timeout=5m; \
	else \
		echo "⚠️ golangci-lint not installed, running basic checks..."; \
		./scripts/lint-check.sh; \
	fi

lint-fix:
	@echo "🔧 Running lint with auto-fix..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --timeout=5m --fix; \
	else \
		echo "⚠️ golangci-lint not installed, running format only..."; \
		make format; \
	fi

check:
	@echo "✅ Running all checks (like CI)..."
	@./scripts/lint-check.sh