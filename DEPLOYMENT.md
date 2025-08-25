# Deployment Guide

This document explains how to deploy nativefire to various platforms.

## Deployment Platforms

### 1. Homebrew (macOS/Linux)
- **Platforms**: macOS, Linux
- **User Command**: `brew tap clix-so/tap && brew install nativefire`
- **Automation**: GoReleaser automatically updates homebrew-tap repository on GitHub releases

### 2. npm/npx (All platforms)
- **Platforms**: Windows, macOS, Linux (any platform with Node.js installed)
- **User Commands**: `npm install -g nativefire` or `npx nativefire@latest`
- **Automation**: GitHub Actions automatically deploys to npm on releases

### 3. Docker (All platforms)
- **Platforms**: Any platform supporting Docker
- **User Command**: `docker run --rm ghcr.io/clix-so/nativefire:latest`
- **Automation**: GoReleaser automatically builds and deploys Docker images

### 4. Direct Download
- **Platforms**: Windows, macOS, Linux
- **User Action**: Direct binary download from GitHub Releases
- **Automation**: GoReleaser automatically generates cross-platform binaries

## Release Process

### 1. Prerequisites

#### GitHub Repository Setup
```bash
# 1. Create Homebrew tap repository
# https://github.com/clix-so/homebrew-tap

# 2. Configure GitHub Secrets
# HOMEBREW_TAP_GITHUB_TOKEN: Access token for homebrew-tap repository
# NPM_TOKEN: Token for npm deployment
```

#### npm Token Creation
```bash
# Login to npm
npm login

# Create automation token
npm token create --type=automation
```

### 2. Release Execution

#### Automatic Release (Recommended)
```bash
# 1. Create version tag and push
git tag v1.0.0
git push origin v1.0.0

# 2. GitHub Actions automatically:
#    - Runs GoReleaser (binaries, Docker images, Homebrew)
#    - Deploys to npm
#    - Generates release notes
```

#### Manual Release (Development/Testing)
```bash
# 1. Install GoReleaser
go install github.com/goreleaser/goreleaser@latest

# 2. Dry run (testing)
goreleaser release --snapshot --clean

# 3. Actual release (requires tag)
goreleaser release --clean
```

### 3. Deployment Verification

#### Homebrew
```bash
brew tap clix-so/tap
brew install nativefire
nativefire --version
```

#### npm
```bash
npm install -g nativefire
# or
npx nativefire@latest --version
```

#### Docker
```bash
docker run --rm ghcr.io/clix-so/nativefire:latest --version
```

## Release Checklist

### Pre-release
- [ ] Verify all tests pass
- [ ] Update CHANGELOG.md
- [ ] Check package.json version
- [ ] Validate GoReleaser configuration (`goreleaser check`)
- [ ] Update README.md installation guide

### Post-release
- [ ] Test Homebrew installation
- [ ] Test npm installation
- [ ] Test Docker image
- [ ] Test direct download for each platform
- [ ] Review and update release notes

## Troubleshooting

### GoReleaser Issues
```bash
# Validate configuration
goreleaser check

# Test with snapshot build
goreleaser build --snapshot --clean

# Check logs
goreleaser release --clean --debug
```

### npm Deployment Issues
```bash
# Check npm login status
npm whoami

# Check token permissions
npm token list

# Test manual deployment
npm publish --dry-run
```

### Docker Issues
```bash
# Test local build
docker build -t nativefire:test .

# Test container execution
docker run --rm nativefire:test --version
```

## Version Management

### Semantic Versioning
- **Major (1.0.0)**: Breaking changes
- **Minor (0.1.0)**: Backward-compatible new features
- **Patch (0.0.1)**: Backward-compatible bug fixes

### Release Branch Strategy
```bash
# Development: develop branch
git checkout develop

# Release preparation: release branch
git checkout -b release/v1.0.0

# After completion: merge to main branch
git checkout main
git merge release/v1.0.0
git tag v1.0.0
git push origin main --tags
```

## Deployment Monitoring

### GitHub Actions
- Release workflow status: `.github/workflows/release.yml`
- CI workflow status: `.github/workflows/ci.yml`

### Deployment Status Check
- **Homebrew**: https://github.com/clix-so/homebrew-tap
- **npm**: https://www.npmjs.com/package/nativefire
- **Docker**: https://github.com/clix-so/nativefire/pkgs/container/nativefire
- **GitHub Releases**: https://github.com/clix-so/nativefire/releases

## Rollback Procedures

### Emergency Rollback
```bash
# 1. Delete release from GitHub
# 2. Deprecate version on npm
npm deprecate nativefire@1.0.0 "Critical bug found, use 0.9.0 instead"

# 3. Remove Docker image tags (manual)
# 4. Revert Homebrew Formula to previous version
```

### Normal Rollback
```bash
# 1. Deploy new release with fixes
git tag v1.0.1
git push origin v1.0.1
```