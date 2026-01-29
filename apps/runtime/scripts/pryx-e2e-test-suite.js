#!/usr/bin/env node
/**
 * Pryx Comprehensive E2E Test Suite
 * Automated testing of all features from installation to chat
 */

const { spawn, execSync } = require('child_process');
const fs = require('fs');
const path = require('path');
const WebSocket = require('ws');
const http = require('http');

// Configuration
const PRYX_HOME = process.env.PRYX_HOME || path.join(require('os').homedir(), '.pryx');
const PRYX_BIN = path.join(PRYX_HOME, 'bin', 'pryx-core');
const RUNTIME_PORT_FILE = path.join(PRYX_HOME, 'runtime.port');

// Test results
const results = {
  passed: [],
  failed: [],
  skipped: []
};

let runtimeProcess = null;
let runtimePort = null;

// Utility functions
function log(msg, type = 'info') {
  const prefix = {
    info: 'ℹ️',
    success: '✅',
    error: '❌',
    warning: '⚠️',
    test: '🧪'
  }[type] || 'ℹ️';
  console.log(`${prefix} ${msg}`);
}

async function sleep(ms) {
  return new Promise(resolve => setTimeout(resolve, ms));
}

async function waitForPort(port, timeout = 10000) {
  const start = Date.now();
  while (Date.now() - start < timeout) {
    try {
      await new Promise((resolve, reject) => {
        const req = http.get(`http://localhost:${port}/health`, (res) => {
          if (res.statusCode === 200) resolve();
          else reject(new Error('Not ready'));
        });
        req.on('error', reject);
        req.setTimeout(1000, () => reject(new Error('Timeout')));
      });
      return true;
    } catch {
      await sleep(500);
    }
  }
  return false;
}

// Test cases
const tests = {
  // Phase 1: Installation & Configuration
  async testBinaryExists() {
    log('Checking pryx-core binary exists...', 'test');
    if (!fs.existsSync(PRYX_BIN)) {
      throw new Error(`Binary not found at ${PRYX_BIN}`);
    }
    log('Binary exists', 'success');
  },

  async testHelpCommand() {
    log('Testing help command...', 'test');
    const output = execSync(`${PRYX_BIN} help`, { encoding: 'utf8' });
    if (!output.includes('Usage:')) {
      throw new Error('Help output missing expected content');
    }
    log('Help command works', 'success');
  },

  async testConfigCommand() {
    log('Testing config command...', 'test');
    const output = execSync(`${PRYX_BIN} config list`, { encoding: 'utf8' });
    if (!output.includes('model_provider')) {
      throw new Error('Config output missing expected content');
    }
    log('Config command works', 'success');
  },

  async testDoctorCommand() {
    log('Testing doctor command...', 'test');
    const output = execSync(`${PRYX_BIN} doctor`, { encoding: 'utf8' });
    if (!output.includes('installation')) {
      throw new Error('Doctor output missing expected content');
    }
    log('Doctor command works', 'success');
  },

  // Phase 2: Runtime Startup
  async testRuntimeStarts() {
    log('Testing runtime startup...', 'test');
    
    if (runtimeProcess) {
      runtimeProcess.kill();
      await sleep(1000);
    }

    runtimeProcess = spawn(PRYX_BIN, [], {
      detached: true,
      stdio: ['ignore', 'pipe', 'pipe']
    });

    runtimeProcess.stdout.pipe(fs.createWriteStream('/tmp/pryx-test-runtime.log'));
    runtimeProcess.stderr.pipe(fs.createWriteStream('/tmp/pryx-test-runtime.log'));

    // Wait for port file
    let attempts = 0;
    while (attempts < 20) {
      await sleep(500);
      if (fs.existsSync(RUNTIME_PORT_FILE)) {
        runtimePort = fs.readFileSync(RUNTIME_PORT_FILE, 'utf8').trim();
        break;
      }
      attempts++;
    }

    if (!runtimePort) {
      throw new Error('Runtime did not start (no port file)');
    }

    // Wait for health endpoint
    const ready = await waitForPort(runtimePort, 10000);
    if (!ready) {
      throw new Error('Runtime health check failed');
    }

    log(`Runtime started on port ${runtimePort}`, 'success');
  },

  async testHealthEndpoint() {
    log('Testing health endpoint...', 'test');
    const response = await new Promise((resolve, reject) => {
      http.get(`http://localhost:${runtimePort}/health`, (res) => {
        let data = '';
        res.on('data', chunk => data += chunk);
        res.on('end', () => resolve({ status: res.statusCode, data }));
      }).on('error', reject);
    });

    if (response.status !== 200) {
      throw new Error(`Health check returned ${response.status}`);
    }
    log('Health endpoint responding', 'success');
  },

  async testSkillsEndpoint() {
    log('Testing skills endpoint...', 'test');
    const response = await new Promise((resolve, reject) => {
      http.get(`http://localhost:${runtimePort}/skills`, (res) => {
        let data = '';
        res.on('data', chunk => data += chunk);
        res.on('end', () => resolve({ status: res.statusCode, data }));
      }).on('error', reject);
    });

    if (response.status !== 200) {
      throw new Error(`Skills endpoint returned ${response.status}`);
    }

    const skills = JSON.parse(response.data);
    if (!skills.skills || skills.skills.length === 0) {
      throw new Error('No skills found');
    }

    log(`Found ${skills.skills.length} skills`, 'success');
  },

  async testMCPToolsEndpoint() {
    log('Testing MCP tools endpoint...', 'test');
    const response = await new Promise((resolve, reject) => {
      http.get(`http://localhost:${runtimePort}/mcp/tools`, (res) => {
        let data = '';
        res.on('data', chunk => data += chunk);
        res.on('end', () => resolve({ status: res.statusCode, data }));
      }).on('error', reject);
    });

    if (response.status !== 200) {
      throw new Error(`MCP tools endpoint returned ${response.status}`);
    }
    log('MCP tools endpoint responding', 'success');
  },

  // Phase 3: WebSocket & Chat
  async testWebSocketConnection() {
    log('Testing WebSocket connection...', 'test');
    
    return new Promise((resolve, reject) => {
      const ws = new WebSocket(`ws://localhost:${runtimePort}/ws?surface=e2e-test`);
      const timeout = setTimeout(() => {
        ws.close();
        reject(new Error('WebSocket connection timeout'));
      }, 10000);

      ws.on('open', () => {
        clearTimeout(timeout);
        ws.close();
        log('WebSocket connected', 'success');
        resolve();
      });

      ws.on('error', (err) => {
        clearTimeout(timeout);
        reject(err);
      });
    });
  },

  async testChatWithGLM() {
    log('Testing chat with GLM provider...', 'test');
    
    return new Promise((resolve, reject) => {
      const ws = new WebSocket(`ws://localhost:${runtimePort}/ws?surface=e2e-test&session_id=e2e-session`);
      let responseReceived = false;

      const timeout = setTimeout(() => {
        ws.close();
        if (!responseReceived) {
          reject(new Error('No chat response received within 30s'));
        }
      }, 30000);

      ws.on('open', () => {
        ws.send(JSON.stringify({
          type: 'chat.send',
          session_id: 'e2e-session',
          payload: { content: 'Say hello in exactly 3 words' }
        }));
      });

      ws.on('message', (data) => {
        const evt = JSON.parse(data);
        if (evt.event === 'session.message' && evt.payload?.content) {
          responseReceived = true;
          clearTimeout(timeout);
          ws.close();
          
          if (evt.payload.content.length > 0) {
            log(`Received response: "${evt.payload.content.substring(0, 50)}..."`, 'success');
            resolve();
          } else {
            reject(new Error('Empty response received'));
          }
        }
      });

      ws.on('error', (err) => {
        clearTimeout(timeout);
        reject(err);
      });
    });
  },

  // Phase 4: CLI Skills
  async testSkillsListCommand() {
    log('Testing skills list CLI...', 'test');
    const output = execSync(`${PRYX_BIN} skills list`, { encoding: 'utf8' });
    if (!output.includes('Available Skills')) {
      throw new Error('Skills list output unexpected');
    }
    log('Skills list CLI works', 'success');
  },

  async testCostSummaryCommand() {
    log('Testing cost summary CLI...', 'test');
    const output = execSync(`${PRYX_BIN} cost summary`, { encoding: 'utf8' });
    if (!output.includes('Total Cost')) {
      throw new Error('Cost summary output unexpected');
    }
    log('Cost summary CLI works', 'success');
  }
};

// Main test runner
async function runAllTests() {
  log('Starting Pryx E2E Test Suite', 'info');
  log(`Pryx Home: ${PRYX_HOME}`, 'info');
  log(`Binary: ${PRYX_BIN}`, 'info');
  console.log('');

  const testOrder = [
    'testBinaryExists',
    'testHelpCommand',
    'testConfigCommand',
    'testDoctorCommand',
    'testRuntimeStarts',
    'testHealthEndpoint',
    'testSkillsEndpoint',
    'testMCPToolsEndpoint',
    'testWebSocketConnection',
    'testChatWithGLM',
    'testSkillsListCommand',
    'testCostSummaryCommand'
  ];

  for (const testName of testOrder) {
    try {
      await tests[testName]();
      results.passed.push(testName);
    } catch (err) {
      log(`Test failed: ${err.message}`, 'error');
      results.failed.push({ name: testName, error: err.message });
    }
    console.log('');
  }

  // Cleanup
  if (runtimeProcess) {
    runtimeProcess.kill();
    log('Runtime stopped', 'info');
  }

  // Report
  console.log('');
  console.log('═'.repeat(60));
  log('E2E TEST RESULTS', 'info');
  console.log('═'.repeat(60));
  log(`Passed: ${results.passed.length}`, 'success');
  log(`Failed: ${results.failed.length}`, results.failed.length > 0 ? 'error' : 'info');
  
  if (results.failed.length > 0) {
    console.log('');
    log('Failed tests:', 'error');
    results.failed.forEach(({ name, error }) => {
      console.log(`  - ${name}: ${error}`);
    });
  }

  console.log('');
  const success = results.failed.length === 0;
  if (success) {
    log('🎉 ALL TESTS PASSED!', 'success');
  } else {
    log(`⚠️ ${results.failed.length} test(s) failed`, 'warning');
  }

  process.exit(success ? 0 : 1);
}

// Handle errors
process.on('unhandledRejection', (err) => {
  log(`Unhandled error: ${err.message}`, 'error');
  if (runtimeProcess) runtimeProcess.kill();
  process.exit(1);
});

process.on('SIGINT', () => {
  log('Interrupted', 'warning');
  if (runtimeProcess) runtimeProcess.kill();
  process.exit(1);
});

// Run
runAllTests();
