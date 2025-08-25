#!/usr/bin/env node

/**
 * NativeFire - Firebase setup made easy for native development
 * 
 * This is the npm package entry point that automatically downloads
 * the appropriate binary for the user's platform and executes it.
 */

const { spawn } = require('child_process');
const path = require('path');
const fs = require('fs');

// Binary paths
const binDir = path.join(__dirname, 'bin');
const binaryName = process.platform === 'win32' ? 'nativefire.exe' : 'nativefire';
const binaryPath = path.join(binDir, binaryName);

async function main() {
  // Check if binary exists
  if (!fs.existsSync(binaryPath)) {
    console.error('❌ NativeFire binary not found!');
    console.error('');
    console.error('This might happen if:');
    console.error('  • The installation was interrupted');
    console.error('  • Your platform is not supported');
    console.error('  • There was a download error');
    console.error('');
    console.error('💡 Try reinstalling:');
    console.error('   npm uninstall -g nativefire');
    console.error('   npm install -g nativefire');
    console.error('');
    console.error('💡 Alternative installation methods:');
    console.error('   • Homebrew: brew install clix-so/tap/nativefire');
    console.error('   • Manual: Download from https://github.com/clix-so/nativefire/releases');
    process.exit(1);
  }

  // Execute the binary with all arguments
  const args = process.argv.slice(2);
  
  const child = spawn(binaryPath, args, {
    stdio: 'inherit',
    windowsHide: false
  });

  // Handle process events
  child.on('error', (error) => {
    console.error('❌ Failed to start nativefire:', error.message);
    process.exit(1);
  });

  child.on('close', (code) => {
    process.exit(code || 0);
  });

  // Handle signals
  process.on('SIGINT', () => {
    child.kill('SIGINT');
  });

  process.on('SIGTERM', () => {
    child.kill('SIGTERM');
  });
}

// Run main function
if (require.main === module) {
  main().catch((error) => {
    console.error('❌ Unexpected error:', error.message);
    process.exit(1);
  });
}

module.exports = { main };