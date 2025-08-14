#!/usr/bin/env node

const fs = require('fs');
const path = require('path');

function uninstall() {
  const binDir = path.join(__dirname, '..', 'bin');
  
  try {
    if (fs.existsSync(binDir)) {
      fs.rmSync(binDir, { recursive: true, force: true });
      console.log('✅ nativefire binary uninstalled successfully!');
    }
  } catch (error) {
    console.error('❌ Uninstallation failed:', error.message);
  }
}

if (require.main === module) {
  uninstall();
}