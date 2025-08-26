# Security Policy

## Supported Versions

We actively support the following versions of NativeFire with security updates:

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | ✅ Active support  |
| < 1.0   | ❌ Not supported   |

## Reporting a Vulnerability

We take security vulnerabilities seriously. If you discover a security vulnerability in NativeFire, please report it to us privately.

### How to Report

1. **GitHub Security Advisories** (Preferred)
   - Go to the [Security Advisories page](https://github.com/clix-so/nativefire/security/advisories)
   - Click "New draft security advisory"
   - Fill out the form with vulnerability details

2. **Email** (Alternative)
   - Send an email to: security@nativefire.dev
   - Include "SECURITY" in the subject line
   - Provide detailed information about the vulnerability

### What to Include

Please include the following information in your report:

- **Description**: A clear description of the vulnerability
- **Steps to Reproduce**: Detailed steps to reproduce the issue
- **Impact**: Potential impact and attack scenarios
- **Affected Versions**: Which versions are affected
- **Proposed Fix**: If you have suggestions for fixing the issue
- **Proof of Concept**: Code or commands demonstrating the vulnerability (if safe to share)

### Response Timeline

We aim to respond to security reports according to the following timeline:

- **Initial Response**: Within 24 hours
- **Assessment**: Within 72 hours
- **Resolution**: Varies based on severity and complexity
  - Critical: 1-7 days
  - High: 7-14 days
  - Medium: 14-30 days
  - Low: 30-90 days

### Security Process

1. **Report Received**: We acknowledge receipt and begin assessment
2. **Verification**: We verify and reproduce the vulnerability
3. **Assessment**: We assess the impact and severity
4. **Development**: We develop and test a fix
5. **Disclosure**: We coordinate responsible disclosure
6. **Release**: We release the security fix
7. **Advisory**: We publish a security advisory (if applicable)

### Severity Guidelines

We use the following severity classifications:

#### Critical
- Remote code execution
- Privilege escalation to system administrator
- Large-scale data exposure

#### High  
- Significant data exposure
- Authentication bypass
- Privilege escalation to application level

#### Medium
- Cross-site scripting (XSS)
- SQL injection with limited impact
- Information disclosure

#### Low
- Minor information disclosure
- Issues requiring significant user interaction
- Theoretical vulnerabilities with no practical exploit

### What We Promise

- We will respond to your report promptly and keep you updated
- We will not take legal action against security researchers who:
  - Report vulnerabilities responsibly and privately
  - Do not access more data than necessary to demonstrate the vulnerability
  - Do not harm our systems or users
  - Give us reasonable time to fix the issue before public disclosure

### Recognition

We appreciate security researchers who help keep NativeFire secure:

- We will acknowledge your contribution in our security advisory (unless you prefer to remain anonymous)
- We may feature your contribution in our release notes
- For significant findings, we may provide recognition on our website

### Security Best Practices for Users

To help protect yourself when using NativeFire:

1. **Keep Updated**: Always use the latest version of NativeFire
2. **Verify Downloads**: Only download NativeFire from official sources
3. **Review Permissions**: Understand what permissions NativeFire requires
4. **Secure Environment**: Use NativeFire in a secure development environment
5. **Firebase Security**: Follow Firebase security best practices for your projects

### Security Features

NativeFire includes several security features:

- **Dependency Validation**: Checks for required external tools before execution
- **Input Sanitization**: Validates user inputs and file paths
- **Minimal Permissions**: Runs with minimal required permissions
- **Secure Defaults**: Uses secure configuration defaults
- **No Secret Storage**: Does not store sensitive information locally

### Known Security Considerations

Users should be aware of the following security considerations:

1. **Firebase CLI Authentication**: NativeFire relies on Firebase CLI authentication
2. **File System Access**: NativeFire modifies project files and directories
3. **Network Requests**: NativeFire makes requests to Firebase APIs
4. **External Dependencies**: NativeFire depends on external tools (Firebase CLI, etc.)

### Security Updates

Security updates are distributed through:

- **GitHub Releases**: Security patches are released as new versions
- **Package Managers**: Updates are available through Homebrew, npm, etc.
- **Security Advisories**: Critical issues are announced via GitHub Security Advisories
- **Release Notes**: Security fixes are documented in release notes

### Contact Information

For security-related questions or concerns:

- **Security Email**: security@nativefire.dev
- **GitHub Security**: https://github.com/clix-so/nativefire/security
- **General Issues**: https://github.com/clix-so/nativefire/issues (for non-security issues only)

---

Thank you for helping keep NativeFire and its users secure! 🔒