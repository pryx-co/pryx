# Credential Storage Testing Guide

## ⚠️ IMPORTANT: Never Store Real Credentials in Code

**DO NOT** commit actual API keys, passwords, or credentials to:
- Source code
- Test files
- Configuration files
- Documentation
- Git history

## Testing with Mock Data

All credential storage tests use **mock/test data only**:

```typescript
// ✅ CORRECT - Use mock data
const TEST_API_KEY = 'sk-test-1234567890abcdef';
const TEST_PASSWORD = 'demo-password-not-real';

// ❌ WRONG - Never do this
const REAL_API_KEY = 'dd358251b05b48c688891384f81a4398.4lq5rmHmYl2WIXrt';
```

## Running the Demonstration

To see how credential storage works:

```bash
# Run the demonstration script
bun run scripts/demo-credential-storage.ts
```

This demonstrates:
1. Vault initialization with Argon2id
2. AES-256-GCM encryption
3. Atomic file writes
4. Secure memory clearing

## Integration Testing

For actual credential storage tests:

1. **Use environment variables** for test credentials:
```bash
export TEST_API_KEY="your-test-key-here"
export TEST_MASTER_PASSWORD="test-password"
```

2. **Never commit the .env file**:
```gitignore
.env
.env.local
*.key
secrets/
```

3. **Use mock services** in tests:
```typescript
// Mock the API client
vi.mock('@pryx/vault', () => ({
  PasswordManager: vi.fn(() => ({
    unlock: vi.fn(),
    encrypt: vi.fn(() => mockEncryptedData),
  })),
}));
```

## Security Checklist

Before committing code that handles credentials:

- [ ] No hardcoded credentials
- [ ] No credential logging
- [ ] Test data only in test files
- [ ] .env files in .gitignore
- [ ] Secrets scanning enabled in CI
- [ ] Pre-commit hooks configured

## If You Accidentally Commit Credentials

1. **Immediately revoke the credential** at the provider
2. **Rotate/generate new credentials**
3. **Remove from Git history**:
   ```bash
   git filter-branch --force --index-filter \
   'git rm --cached --ignore-unmatch path/to/file' \
   HEAD
   ```
4. **Force push** (if not pushed to shared branches)
5. **Audit access logs** for the exposed credential

## Test Credential Format

When testing, use this format:
```
sk-test-[timestamp]-[random]
```

Example:
```
sk-test-20260130-12345678abcdef
```

This makes it obvious the credential is for testing only.

## Vault Storage Verification

To verify your vault is working:

```typescript
import { PasswordManager } from '@pryx/vault';

const manager = new PasswordManager();
await manager.unlock('your-master-password');

// Test encryption
const testData = Buffer.from('test-data');
const encrypted = await manager.encrypt(testData);
const decrypted = await manager.decrypt(encrypted);

console.log('Vault working:', decrypted.toString() === 'test-data');
```

## Production Deployment

For production:

1. Use hardware security modules (HSM) if available
2. Enable audit logging
3. Set up monitoring for vault access
4. Implement key rotation policies
5. Use environment-specific vaults
6. Enable MFA for vault access

## Support

If you need help with credential storage:
- Review the demo script: `scripts/demo-credential-storage.ts`
- Check vault tests: `packages/vault/tests/`
- Read the security guide: `docs/security/vault-usage.md`
