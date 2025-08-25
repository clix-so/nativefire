# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Multi-platform deployment support (Homebrew, npm, Docker)
- Comprehensive dependency checking system
- Enhanced error handling with user-friendly installation guides
- Automated CI/CD pipeline with GitHub Actions
- Cross-platform binary releases via GoReleaser

### Changed
- Improved dependency validation with detailed missing dependency reports
- Enhanced install scripts for npm package with automatic platform detection
- Updated project structure for better maintainability

### Fixed
- Dependency checks now provide clear installation instructions
- Improved error messages when external CLI tools are missing

## [1.0.0] - 2024-XX-XX

### Added
- Initial release of nativefire CLI
- Firebase project and platform auto-detection
- Support for Android, iOS, macOS, Windows, and Linux platforms
- Automatic Firebase app registration and configuration file download
- Smart configuration file placement based on project structure
- Firebase initialization code injection for supported platforms
- Interactive project selection
- Comprehensive CLI with help documentation
- Verbose logging support
- Configuration file support (.nativefire.yaml)
- Bundle ID and Package Name auto-detection
- CocoaPods and Swift Package Manager support for iOS
- Gradle integration for Android projects
- Cross-platform compatibility

### Features
- **Multi-platform Detection**: Automatically detects Android, iOS, macOS, Windows, and Linux projects
- **Firebase Integration**: Seamlessly integrates with Firebase CLI for app registration and configuration
- **Smart Placement**: Places configuration files in correct directories for each platform
- **Code Injection**: Automatically adds Firebase initialization code to project entry points
- **Package Manager Support**: Handles CocoaPods, SPM, and Gradle dependencies
- **Interactive CLI**: User-friendly command-line interface with comprehensive help
- **Verbose Output**: Detailed logging for debugging and troubleshooting

### Supported Platforms
- **Android**: Gradle-based projects with automatic google-services.json placement
- **iOS**: Xcode projects with Podfile or SPM support, GoogleService-Info.plist placement
- **macOS**: Xcode projects with Firebase configuration
- **Windows**: CMake and Visual Studio project support
- **Linux**: CMake and Makefile project support

### Installation Methods
- **Homebrew**: `brew tap clix-so/tap && brew install nativefire`
- **npm/npx**: `npm install -g nativefire` or `npx nativefire@latest`
- **Docker**: `docker run --rm ghcr.io/clix-so/nativefire:latest`
- **Direct Download**: Platform-specific binaries from GitHub Releases

---

## Release Notes Format

### Types of Changes
- `Added` for new features
- `Changed` for changes in existing functionality
- `Deprecated` for soon-to-be removed features
- `Removed` for now removed features
- `Fixed` for any bug fixes
- `Security` for vulnerability fixes