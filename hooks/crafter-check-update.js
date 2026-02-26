#!/usr/bin/env node
const fs = require('fs');
const path = require('path');
const os = require('os');
const { spawn } = require('child_process');

const homeDir = os.homedir();
const cwd = process.cwd();
const cacheDir = path.join(homeDir, '.claude', 'cache');
const cacheFile = path.join(cacheDir, 'crafter-update-check.json');

const globalVersionFile = path.join(homeDir, '.claude', 'crafter', 'VERSION');
const projectVersionFile = path.join(cwd, '.claude', 'crafter', 'VERSION');

// Exit silently if crafter is not installed
if (!fs.existsSync(globalVersionFile) && !fs.existsSync(projectVersionFile)) {
  process.exit(0);
}

// Ensure cache directory exists
try {
  if (!fs.existsSync(cacheDir)) {
    fs.mkdirSync(cacheDir, { recursive: true });
  }
} catch (e) {}

// Read current installed version
let installedVersion = null;
try {
  if (fs.existsSync(globalVersionFile)) {
    installedVersion = fs.readFileSync(globalVersionFile, 'utf8').trim();
  } else if (fs.existsSync(projectVersionFile)) {
    installedVersion = fs.readFileSync(projectVersionFile, 'utf8').trim();
  }
} catch (e) {}

// Synchronous part: read cache and print notice if update is available
try {
  const raw = fs.readFileSync(cacheFile, 'utf8');
  const cache = JSON.parse(raw);
  if (cache && cache.update_available === true && installedVersion === cache.installed) {
    process.stdout.write(
      `Note: Crafter update available (installed: ${cache.installed}, latest: ${cache.latest}). Run: curl -fsSL https://raw.githubusercontent.com/richardriman/crafter/main/install.sh | bash\n`
    );
  }
} catch (e) {
  // Cache doesn't exist, is invalid, or unreadable — print nothing
}

// Background part: check GitHub for latest release and update cache
const bgScript = `
  const fs = require('fs');
  const { spawnSync } = require('child_process');

  const cacheFile = ${JSON.stringify(cacheFile)};
  const projectVersionFile = ${JSON.stringify(projectVersionFile)};
  const globalVersionFile = ${JSON.stringify(globalVersionFile)};

  try {
    // Read installed version
    let installed = '0.0.0';
    if (fs.existsSync(globalVersionFile)) {
      installed = fs.readFileSync(globalVersionFile, 'utf8').trim();
    } else if (fs.existsSync(projectVersionFile)) {
      installed = fs.readFileSync(projectVersionFile, 'utf8').trim();
    }

    // Check cache freshness — skip GitHub call if checked within 24 hours
    // Also invalidate cache if installed version changed (e.g. after upgrade)
    const now = Math.floor(Date.now() / 1000);
    try {
      const raw = fs.readFileSync(cacheFile, 'utf8');
      const cache = JSON.parse(raw);
      if (cache && cache.checked && (now - cache.checked) < 14400 && cache.installed === installed) {
        process.exit(0);
      }
    } catch (e) {
      // Cache missing or invalid — proceed with GitHub check
    }

    // Fetch latest release from GitHub
    let latest = null;
    try {
      const result = spawnSync(
        'curl',
        ['-sf', '--max-time', '5', 'https://api.github.com/repos/richardriman/crafter/releases/latest'],
        { encoding: 'utf8', timeout: 10000, windowsHide: true }
      );
      if (result.status === 0 && result.stdout) {
        const data = JSON.parse(result.stdout);
        if (data && data.tag_name) {
          latest = data.tag_name.replace(/^v/, '');
        }
      }
    } catch (e) {}

    if (latest === null) {
      process.exit(0);
    }

    // Compare semver: returns true only when latest is strictly newer than installed
    function isNewer(latest, installed) {
      const a = latest.split('.').map(Number);
      const b = installed.split('.').map(Number);
      for (let i = 0; i < Math.max(a.length, b.length); i++) {
        const av = a[i] || 0;
        const bv = b[i] || 0;
        if (av > bv) return true;
        if (av < bv) return false;
      }
      return false;
    }

    // Write updated cache
    const cacheResult = {
      update_available: isNewer(latest, installed),
      installed,
      latest,
      checked: now
    };

    fs.writeFileSync(cacheFile, JSON.stringify(cacheResult));
  } catch (e) {}
`;

const child = spawn(process.execPath, ['-e', bgScript], {
  stdio: 'ignore',
  windowsHide: true,
  detached: true
});

child.unref();
