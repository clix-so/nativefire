#!/usr/bin/env node

const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');
const https = require('https');

const VERSION = '1.0.0';
const GITHUB_REPO = 'clix-so/nativefire';

function getPlatform() {
  const platform = process.platform;
  const arch = process.arch;

  let platformName;
  let archName;

  switch (platform) {
    case 'darwin':
      platformName = 'darwin';
      break;
    case 'linux':
      platformName = 'linux';
      break;
    case 'win32':
      platformName = 'windows';
      break;
    default:
      throw new Error(`Unsupported platform: ${platform}`);
  }

  switch (arch) {
    case 'x64':
      archName = 'amd64';
      break;
    case 'arm64':
      archName = 'arm64';
      break;
    default:
      throw new Error(`Unsupported architecture: ${arch}`);
  }

  return { platform: platformName, arch: archName };
}

function downloadBinary() {
  const { platform, arch } = getPlatform();
  const extension = platform === 'windows' ? '.exe' : '';
  const binaryName = `nativefire-${platform}-${arch}${extension}`;
  const downloadUrl = `https://github.com/${GITHUB_REPO}/releases/download/v${VERSION}/${binaryName}`;
  
  const binDir = path.join(__dirname, '..', 'bin');
  const binaryPath = path.join(binDir, `nativefire${extension}`);

  if (!fs.existsSync(binDir)) {
    fs.mkdirSync(binDir, { recursive: true });
  }

  console.log(`Downloading nativefire binary for ${platform}-${arch}...`);
  console.log(`URL: ${downloadUrl}`);

  return new Promise((resolve, reject) => {
    const file = fs.createWriteStream(binaryPath);
    
    https.get(downloadUrl, (response) => {
      if (response.statusCode === 302 || response.statusCode === 301) {
        // Handle redirect
        https.get(response.headers.location, (redirectResponse) => {
          if (redirectResponse.statusCode === 200) {
            redirectResponse.pipe(file);
            file.on('finish', () => {
              file.close();
              fs.chmodSync(binaryPath, '755');
              console.log('✅ nativefire binary installed successfully!');
              resolve();
            });
          } else {
            reject(new Error(`Failed to download: ${redirectResponse.statusCode}`));
          }
        });
      } else if (response.statusCode === 200) {
        response.pipe(file);
        file.on('finish', () => {
          file.close();
          fs.chmodSync(binaryPath, '755');
          console.log('✅ nativefire binary installed successfully!');
          resolve();
        });
      } else {
        reject(new Error(`Failed to download: ${response.statusCode}`));
      }
    }).on('error', (err) => {
      reject(err);
    });
  });
}

async function install() {
  try {
    await downloadBinary();
    console.log('Installation completed successfully!');
    console.log('You can now use: npx nativefire --help');
  } catch (error) {
    console.error('❌ Installation failed:', error.message);
    console.log('');
    console.log('Alternative installation methods:');
    console.log('1. Homebrew (macOS): brew install nativefire');
    console.log('2. Download from: https://github.com/clix-so/nativefire/releases');
    process.exit(1);
  }
}

if (require.main === module) {
  install();
}