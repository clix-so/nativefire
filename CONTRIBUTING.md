# Contributing to NativeFire

Thank you for your interest in contributing to NativeFire! This document provides guidelines and information for contributors.

## Table of Contents
- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Contributing Process](#contributing-process)
- [Coding Standards](#coding-standards)
- [Testing](#testing)
- [Documentation](#documentation)
- [Release Process](#release-process)

## Code of Conduct

This project adheres to a code of conduct. By participating, you are expected to uphold this code. Please report unacceptable behavior to the project maintainers.

### Our Pledge
- Be respectful and inclusive
- Welcome newcomers and help them learn
- Focus on constructive feedback
- Prioritize the community's well-being

## Getting Started

### Prerequisites
- Go 1.21 or later
- Git
- Firebase CLI (for testing)
- Node.js (for npm package testing)

### Repository Structure
```
nativefire/
├── cmd/                 # CLI commands and root command
├── internal/           # Internal packages
│   ├── dependencies/   # Dependency management
│   ├── firebase/       # Firebase integration
│   ├── platform/       # Platform detection and handling
│   └── ui/            # User interface utilities
├── scripts/            # Build and deployment scripts
├── testdata/          # Test data and fixtures
├── .github/           # GitHub Actions workflows
└── docs/              # Documentation
```

## Development Setup

### 1. Fork and Clone
```bash
# Fork the repository on GitHub, then:
git clone https://github.com/YOUR_USERNAME/nativefire.git
cd nativefire
```

### 2. Install Dependencies
```bash
# Install Go dependencies
go mod download

# Install development tools (optional but recommended)
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### 3. Build and Test
```bash
# Build the project
go build -o nativefire

# Run tests
go test ./...

# Run with verbose output
go test -v ./...
```

### 4. Development Commands
```bash
# Format code
go fmt ./...

# Run linter (if installed)
golangci-lint run

# Build for current platform
go build -o nativefire

# Test the binary
./nativefire --help
```

## Contributing Process

### 1. Create an Issue
Before starting work, create an issue to discuss:
- Bug reports with reproduction steps
- Feature requests with use cases
- Documentation improvements
- Performance enhancements

### 2. Create a Branch
```bash
# Create a feature branch from main
git checkout main
git pull origin main
git checkout -b feature/your-feature-name

# Or for bug fixes
git checkout -b fix/issue-description
```

### 3. Make Changes
- Follow the coding standards below
- Add tests for new functionality
- Update documentation as needed
- Ensure all tests pass

### 4. Commit Changes
```bash
# Stage your changes
git add .

# Commit with descriptive message
git commit -m "feat: add support for Flutter projects

- Add Flutter project detection
- Update configuration placement logic
- Add tests for Flutter platform detection

Closes #123"
```

### 5. Push and Create PR
```bash
# Push your branch
git push origin feature/your-feature-name

# Create a Pull Request on GitHub with:
# - Clear description of changes
# - Reference to related issues
# - Screenshots (if UI changes)
# - Testing instructions
```

## Coding Standards

### Go Style Guide
- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `go fmt` for formatting
- Use meaningful variable and function names
- Add comments for exported functions and types
- Keep functions focused and small

### Code Structure
```go
// Package comment
package main

import (
    // Standard library imports
    "fmt"
    "os"
    
    // Third-party imports
    "github.com/spf13/cobra"
    
    // Local imports
    "github.com/clix-so/nativefire/internal/platform"
)

// Exported function with comment
func ExampleFunction(param string) error {
    // Implementation
    return nil
}
```

### Error Handling
- Always handle errors explicitly
- Provide meaningful error messages
- Use fmt.Errorf for error wrapping
- Include context in error messages

```go
// Good error handling
if err := someOperation(); err != nil {
    return fmt.Errorf("failed to perform operation: %w", err)
}
```

## Testing

### Unit Tests
- Write tests for all new functionality
- Use table-driven tests when appropriate
- Test both success and error cases
- Aim for good test coverage

```go
func TestPlatformDetection(t *testing.T) {
    tests := []struct {
        name     string
        files    []string
        expected platform.Type
    }{
        {
            name:     "Android project",
            files:    []string{"build.gradle", "app/build.gradle"},
            expected: platform.Android,
        },
        // More test cases...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Integration Tests
- Test CLI commands end-to-end
- Use testdata directory for fixtures
- Mock external dependencies when possible

### Test Commands
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test
go test -run TestPlatformDetection ./internal/platform/

# Run tests with race detection
go test -race ./...
```

## Documentation

### Code Documentation
- Add comments for exported functions, types, and constants
- Use godoc format for documentation
- Include examples in comments when helpful

```go
// DetectPlatform automatically detects the current project's platform
// based on files and directory structure in the current working directory.
// It returns the detected platform type or an error if no platform is detected.
//
// Example:
//   platform, err := DetectPlatform()
//   if err != nil {
//       log.Fatal(err)
//   }
//   fmt.Printf("Detected platform: %s\n", platform.Name())
func DetectPlatform() (Platform, error) {
    // Implementation
}
```

### README Updates
- Update installation instructions for new features
- Add examples for new commands or options
- Update troubleshooting section as needed

### Changelog
- Add entries to CHANGELOG.md for all changes
- Follow the established format
- Include breaking changes in the appropriate section

## Release Process

### Version Numbering
We follow [Semantic Versioning](https://semver.org/):
- **MAJOR**: Breaking changes
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes (backward compatible)

### Release Checklist
1. Update CHANGELOG.md
2. Update version in package.json
3. Test all installation methods
4. Create and push version tag
5. Verify automated deployment
6. Test deployed packages

## Questions and Support

### Getting Help
- Check existing issues and discussions
- Read the documentation thoroughly
- Ask questions in GitHub Discussions
- Contact maintainers for sensitive issues

### Reporting Bugs
Include in your bug report:
- Operating system and version
- Go version
- NativeFire version
- Steps to reproduce
- Expected vs actual behavior
- Relevant logs or error messages

### Feature Requests
Include in your feature request:
- Use case description
- Proposed solution
- Alternative solutions considered
- Additional context

## Recognition

Contributors will be recognized in:
- Release notes for significant contributions
- README contributors section
- Special mentions for major features

Thank you for contributing to NativeFire! 🔥