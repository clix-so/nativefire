#!/usr/bin/env node

const os = require('os');
const fs = require('fs');
const path = require('path');
const https = require('https');
const { execSync } = require('child_process');

// GitHub repository info
const REPO_OWNER = 'clix-so';
const REPO_NAME = 'nativefire';
const BINARY_NAME = 'nativefire';

// Platform and architecture mapping for GoReleaser
function getPlatformInfo() {
  const platform = os.platform();
  const arch = os.arch();
  
  let osName, archName;
  
  // Map Node.js platform names to GoReleaser names
  switch (platform) {
    case 'darwin':
      osName = 'Darwin';
      break;
    case 'linux':
      osName = 'Linux';
      break;
    case 'win32':
      osName = 'Windows';
      break;
    default:
      throw new Error(`Unsupported platform: ${platform}`);
  }
  
  // Map Node.js arch names to GoReleaser names
  switch (arch) {
    case 'x64':
      archName = 'x86_64';
      break;
    case 'arm64':
      archName = 'arm64';
      break;
    case 'ia32':
      archName = 'i386';
      break;
    default:
      throw new Error(`Unsupported architecture: ${arch}`);
  }
  
  return { osName, archName, platform };
}

// Get the latest release info from GitHub API
async function getLatestRelease() {
  return new Promise((resolve, reject) => {
    const options = {
      hostname: 'api.github.com',
      path: `/repos/${REPO_OWNER}/${REPO_NAME}/releases/latest`,
      headers: {
        'User-Agent': 'nativefire-npm-installer'
      }
    };
    
    https.get(options, (res) => {
      let data = '';
      res.on('data', (chunk) => data += chunk);
      res.on('end', () => {
        try {
          resolve(JSON.parse(data));
        } catch (error) {
          reject(error);
        }
      });
    }).on('error', reject);
  });
}

// Download binary from GitHub releases
async function downloadBinary(url, targetPath) {
  return new Promise((resolve, reject) => {
    const file = fs.createWriteStream(targetPath);
    
    https.get(url, (response) => {
      if (response.statusCode === 302 || response.statusCode === 301) {
        // Follow redirect
        return downloadBinary(response.headers.location, targetPath);
      }
      
      if (response.statusCode !== 200) {
        reject(new Error(`Failed to download: ${response.statusCode}`));
        return;
      }
      
      response.pipe(file);
      file.on('finish', () => {
        file.close();
        // Make binary executable on Unix systems
        if (process.platform !== 'win32') {
          fs.chmodSync(targetPath, 0o755);
        }
        resolve();
      });
    }).on('error', (error) => {
      fs.unlink(targetPath, () => {}); // Delete the file on error
      reject(error);
    });
  });
}

// Extract archive (for .tar.gz and .zip files)
function extractArchive(archivePath, targetDir) {
  const ext = path.extname(archivePath);
  
  if (ext === '.gz') {
    // Handle .tar.gz files
    execSync(`tar -xzf "${archivePath}" -C "${targetDir}"`, { stdio: 'inherit' });
  } else if (ext === '.zip') {
    // Handle .zip files
    if (process.platform === 'win32') {
      execSync(`powershell -command "Expand-Archive -Path '${archivePath}' -DestinationPath '${targetDir}'"`, { stdio: 'inherit' });
    } else {
      execSync(`unzip -q "${archivePath}" -d "${targetDir}"`, { stdio: 'inherit' });
    }
  }
}

async function install() {
  try {
    console.log('🔍 Installing nativefire binary...');
    
    const { osName, archName, platform } = getPlatformInfo();
    console.log(`📋 Detected platform: ${osName}_${archName}`);
    
    // Get latest release info
    console.log('📡 Fetching latest release information...');
    const release = await getLatestRelease();
    
    // Find the appropriate asset
    const extension = platform === 'win32' ? '.zip' : '.tar.gz';
    const assetName = `nativefire_${osName}_${archName}${extension}`;
    
    const asset = release.assets.find(asset => asset.name === assetName);
    if (!asset) {
      throw new Error(`No binary found for platform ${osName}_${archName}`);
    }
    
    console.log(`📦 Downloading ${asset.name}...`);
    
    // Create bin directory
    const binDir = path.join(__dirname, '..', 'bin');
    if (!fs.existsSync(binDir)) {
      fs.mkdirSync(binDir, { recursive: true });
    }
    
    // Download archive
    const archivePath = path.join(binDir, asset.name);
    await downloadBinary(asset.browser_download_url, archivePath);
    
    // Extract binary
    console.log('📂 Extracting binary...');
    extractArchive(archivePath, binDir);
    
    // Clean up archive
    fs.unlinkSync(archivePath);
    
    // Verify binary exists
    const binaryName = platform === 'win32' ? `${BINARY_NAME}.exe` : BINARY_NAME;
    const binaryPath = path.join(binDir, binaryName);
    
    if (!fs.existsSync(binaryPath)) {
      throw new Error(`Binary not found after extraction: ${binaryPath}`);
    }
    
    console.log(`✅ Successfully installed nativefire ${release.tag_name}`);
    console.log(`📍 Binary location: ${binaryPath}`);
    console.log('');
    console.log('🚀 Try running: npx nativefire --help');
    
  } catch (error) {
    console.error('❌ Installation failed:', error.message);
    console.error('');
    console.error('💡 Alternative installation methods:');
    console.error('   • Homebrew: brew install clix-so/tap/nativefire');
    console.error('   • Manual: Download from https://github.com/clix-so/nativefire/releases');
    process.exit(1);
  }
}

// Only run if this script is executed directly
if (require.main === module) {
  install();
}

module.exports = { install };