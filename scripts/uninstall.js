#!/usr/bin/env node

const fs = require('fs');
const path = require('path');

function uninstall() {
  console.log('🗑️  Uninstalling nativefire binary...');
  
  const binDir = path.join(__dirname, '..', 'bin');
  
  try {
    if (fs.existsSync(binDir)) {
      fs.rmSync(binDir, { recursive: true, force: true });
      console.log('✅ nativefire binary uninstalled successfully!');
      console.log('📍 Removed directory:', binDir);
    } else {
      console.log('💭 No binary found to uninstall');
    }
  } catch (error) {
    console.error('❌ Uninstallation failed:', error.message);
    process.exit(1);
  }
}

// Only run if this script is executed directly
if (require.main === module) {
  uninstall();
}

module.exports = { uninstall };