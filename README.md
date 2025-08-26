# NativeFire

🔥 Simplify Firebase setup in native development environments

NativeFire is a CLI tool that automatically detects your native development environment and sets up Firebase configuration for Android, iOS, macOS, Windows, and Linux platforms. Similar to how `flutterfire` works for Flutter, but for native applications.

## Features

- 🎯 **Auto-detection**: Automatically detects your platform (Android, iOS, macOS, Windows, Linux)
- 🔧 **Firebase Integration**: Uses Firebase CLI to register apps and download configurations
- 📁 **Smart Placement**: Places configuration files in the correct directories for each platform
- 🚀 **Code Injection**: Automatically adds Firebase initialization code to your app's entry point
- 📝 **Verbose Logging**: Clear error messages and detailed output when enabled
- 🏠 **Multiple Installation Methods**: Install via Homebrew or npm/npx

## Installation

### Homebrew (macOS/Linux) - Recommended
```bash
# Install directly (automatically adds tap)
brew install clix-so/homebrew-nativefire/nativefire

# Or add tap first, then install
brew tap clix-so/homebrew-nativefire
brew install nativefire
```

### npm/npx - Cross-platform
```bash
# Install globally
npm install -g nativefire

# Or use with npx (no installation required)
npx nativefire@latest configure --help

# Use latest version with npx
npx nativefire@latest configure --auto-detect
```

### Docker (All platforms)
```bash
# Run directly with Docker
docker run --rm -v $(pwd):/workspace ghcr.io/clix-so/nativefire:latest --help

# Use as alias
alias nativefire='docker run --rm -v $(pwd):/workspace ghcr.io/clix-so/nativefire:latest'
nativefire configure --auto-detect
```

### Manual Installation
Download the latest binary from the [releases page](https://github.com/clix-so/nativefire/releases).

## Prerequisites

- [Firebase CLI](https://firebase.google.com/docs/cli) must be installed and authenticated
  ```bash
  npm install -g firebase-tools
  firebase login
  ```
- A Firebase project created in the [Firebase Console](https://console.firebase.google.com/)

## Usage

### Basic Usage

```bash
# Easiest: Auto-select project and auto-detect platform
nativefire configure --auto-detect

# Select project interactively, auto-detect platform
nativefire configure --auto-detect

# Use specific project, auto-detect platform
nativefire configure --project your-firebase-project-id --auto-detect

# Specify both project and platform explicitly
nativefire configure --project your-firebase-project-id --platform android

# Use verbose output for detailed logging
nativefire configure --auto-detect --verbose
```

### Project Management

```bash
# List all available Firebase projects
nativefire projects list

# Interactively select a Firebase project
nativefire projects select

# Select project and set as default for Firebase CLI
nativefire projects select --use
```

### Command Options

#### Configure Command
- `--project, -p`: Firebase project ID (will prompt if not provided)
- `--platform`: Target platform (android, ios, macos, windows, linux)
- `--auto-detect`: Automatically detect the platform
- `--app-id`: Firebase app ID (optional, will create new app if not provided)
- `--bundle-id`: iOS Bundle ID (will auto-detect or generate if not provided)
- `--package-name`: Android Package Name (will auto-detect or generate if not provided)
- `--verbose, -v`: Enable verbose output
- `--config`: Config file path (default is $HOME/.nativefire.yaml)

#### Projects Commands
- `nativefire projects list`: List all available Firebase projects
- `nativefire projects select`: Interactively select a project
- `nativefire projects select --use`: Select project and set as Firebase CLI default

### Examples

#### Android Project
```bash
cd /path/to/android/project
nativefire configure --auto-detect  # Will prompt for project selection
```

#### iOS Project
```bash
cd /path/to/ios/project
nativefire configure --project my-firebase-project --auto-detect
```

#### Full Interactive Setup
```bash
cd /path/to/your/project
nativefire configure --auto-detect --verbose  # Interactive project selection + platform detection
```

#### Project Management
```bash
# See all your Firebase projects
nativefire projects list

# Pick a project interactively
nativefire projects select

# Use with existing project ID
nativefire configure --project my-existing-project --auto-detect
```

## Platform Detection

NativeFire automatically detects platforms based on project structure:

### Android
- Looks for `build.gradle`, `settings.gradle`
- Searches for `app/build.gradle`, `android/build.gradle`

### iOS
- Looks for `*.xcodeproj`, `*.xcworkspace`
- Searches for `Podfile`, `ios/` directory

### macOS
- Looks for `macos/` directory
- Searches for Xcode projects with Podfile

### Windows
- Looks for `windows/` directory
- Searches for `*.vcxproj`, `*.sln`, `CMakeLists.txt`

### Linux
- Looks for `linux/` directory
- Searches for `CMakeLists.txt`, `Makefile`

## What NativeFire Does

### For Android:
1. 📱 Registers Android app with Firebase
2. 📥 Downloads `google-services.json`
3. 📂 Places config file in `app/` or `android/app/` directory
4. 🔧 Adds Google Services plugin to `build.gradle`
5. 📝 Adds Firebase initialization to MainActivity

### For iOS:
1. 📱 Registers iOS app with Firebase
2. 📥 Downloads `GoogleService-Info.plist`
3. 📂 Places config file in appropriate iOS directory
4. 🔧 Adds Firebase pods to `Podfile`
5. 📝 Adds Firebase initialization to AppDelegate

### For macOS/Windows/Linux:
1. 📱 Registers app with Firebase
2. 📥 Downloads appropriate configuration file
3. 📂 Places config file in platform directory
4. ⚠️ Provides manual setup instructions

## Configuration File

You can create a `.nativefire.yaml` file in your home directory for default settings:

```yaml
project: your-default-firebase-project
verbose: true
```

## App Identifier Detection

NativeFire automatically detects and handles platform-specific app identifiers:

### For Android apps, nativefire will:
- Auto-detect package name from `applicationId` in `build.gradle`
- Auto-detect from `package` attribute in `AndroidManifest.xml`
- Generate a default package name if none found (e.g., `com.firebase.your-project`)
- Accept explicit `--package-name` flag

### For iOS apps, nativefire will:
- Auto-detect Bundle ID from `CFBundleIdentifier` in `Info.plist`
- Auto-detect from `PRODUCT_BUNDLE_IDENTIFIER` in Xcode project
- Generate a default Bundle ID if none found (e.g., `com.firebase.your-project`)
- Accept explicit `--bundle-id` flag

### Examples with explicit identifiers:
```bash
# Android with specific package name
nativefire configure --platform android --package-name com.mycompany.myapp

# iOS with specific bundle ID
nativefire configure --platform ios --bundle-id com.mycompany.myapp
```

## Troubleshooting

### Firebase CLI Not Found
```bash
npm install -g firebase-tools
```

### Not Authenticated with Firebase
```bash
firebase login
```

### Platform Not Detected
Use the `--platform` flag to specify explicitly:
```bash
nativefire configure --project your-project --platform android
```

### Bundle ID or Package Name Cannot Be Empty
If you get this error, nativefire couldn't auto-detect your app identifier. You can:

1. **Specify explicitly:**
   ```bash
   # For iOS
   nativefire configure --bundle-id com.yourcompany.yourapp --auto-detect
   
   # For Android  
   nativefire configure --package-name com.yourcompany.yourapp --auto-detect
   ```

2. **Add to your project files:**
   - Android: Add `applicationId "com.yourcompany.yourapp"` to your `build.gradle`
   - iOS: Set `CFBundleIdentifier` in your `Info.plist`

### Verbose Output for Debugging
```bash
nativefire configure --project your-project --auto-detect --verbose
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Development

### Building from Source

```bash
# Clone the repository
git clone https://github.com/clix-so/nativefire.git
cd nativefire

# Build the binary
make build

# Install locally
make install

# Build for all platforms
make build-all
```

### Running Tests

```bash
make test
```

### Code Quality & Linting

NativeFire uses comprehensive linting and formatting to maintain code quality:

```bash
# Format code
make format

# Run all quality checks (formatting, vet, build, tests)
make check

# Run linting (requires golangci-lint)
make lint

# Run linting with auto-fix
make lint-fix
```

#### Installing golangci-lint (optional)
```bash
# macOS
brew install golangci-lint

# Linux/Windows
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.55.0
```

The GitHub Actions CI automatically runs linting checks. You can run the same checks locally with `make check`.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Inspired by [FlutterFire CLI](https://firebase.flutter.dev/docs/cli/)
- Built with [Cobra](https://github.com/spf13/cobra) for CLI framework
- Uses [Viper](https://github.com/spf13/viper) for configuration management

## Support

- 🐛 [Report a bug](https://github.com/clix-so/nativefire/issues)
- 💡 [Request a feature](https://github.com/clix-so/nativefire/issues)
- 📚 [Documentation](https://github.com/clix-so/nativefire/wiki)

---

Made with ❤️ by the NativeFire team