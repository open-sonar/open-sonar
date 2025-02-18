#!/usr/bin/env node
const path = require('path');
const { spawn } = require('child_process');

function runFosum() {
  const binaryPath = path.join(__dirname, 'bin', 'fosum');
  const child = spawn(binaryPath, process.argv.slice(2), { stdio: 'inherit' });
  child.on('exit', function(code) {
    process.exit(code);
  });
}

runFosum();
