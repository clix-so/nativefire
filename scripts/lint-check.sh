#!/bin/bash

# Local lint check script to validate GitHub Actions will pass
set -e

echo "🔍 Running local lint checks..."

# Check formatting
echo "📝 Checking code formatting..."
UNFORMATTED_FILES=$(gofmt -l . || true)
if [ -n "$UNFORMATTED_FILES" ]; then
    echo "❌ The following files are not properly formatted:"
    echo "$UNFORMATTED_FILES"
    echo "Run: gofmt -w ."
    exit 1
fi
echo "✅ Code formatting looks good!"

# Check go vet
echo "🔍 Running go vet..."
go vet ./...
echo "✅ Go vet passed!"

# Check go mod tidy
echo "📦 Checking go mod tidy..."
go mod tidy
if [ -n "$(git diff go.mod go.sum)" ]; then
    echo "❌ go.mod or go.sum is not tidy. Run: go mod tidy"
    exit 1
fi
echo "✅ Go modules are tidy!"

# Check for common issues
echo "🔧 Running basic checks..."

# Check for TODO/FIXME comments in main code (not tests)
TODO_COUNT=$(grep -r "TODO\|FIXME" --include="*.go" --exclude-dir=testdata . | grep -v "_test.go" | wc -l || true)
if [ "$TODO_COUNT" -gt 0 ]; then
    echo "⚠️  Found $TODO_COUNT TODO/FIXME comments in main code"
    grep -r "TODO\|FIXME" --include="*.go" --exclude-dir=testdata . | grep -v "_test.go" || true
fi

# Check build
echo "🏗️  Checking build..."
go build -v ./...
echo "✅ Build successful!"

# Check tests
echo "🧪 Running tests..."
go test -short ./...
echo "✅ Tests passed!"

echo ""
echo "🎉 All local checks passed!"
echo "💡 GitHub Actions lint should pass when you push."
echo ""
echo "Note: This script runs basic checks. The full golangci-lint"
echo "      may catch additional issues in CI."