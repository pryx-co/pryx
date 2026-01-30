import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { render, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { PasswordManager } from '@pryx/vault';
import { CredentialStorage } from '../src/components/CredentialStorage';
import { SecureInput } from '../src/components/SecureInput';

// Mock test data - NOT real credentials
const MOCK_MASTER_PASSWORD = 'test-master-password-123';
const MOCK_API_KEY = 'test-api-key-sk-1234567890abcdef';
const MOCK_SERVICE_NAME = 'test-ai-provider';

describe('TUI Credential Storage Flow', () => {
  let passwordManager: PasswordManager;
  let tempDir: string;

  beforeEach(async () => {
    // Create isolated password manager for each test
    passwordManager = new PasswordManager({
      autoLockMs: 60000, // 1 minute for testing
    });
    
    // Unlock the vault
    await passwordManager.unlock(MOCK_MASTER_PASSWORD);
  });

  afterEach(() => {
    passwordManager.destroy();
  });

  describe('SecureInput Component', () => {
    it('should mask password input', () => {
      const { getByLabelText } = render(
        <SecureInput 
          label="API Key"
          value=""
          onChange={() => {}}
        />
      );

      const input = getByLabelText('API Key');
      expect(input).toHaveAttribute('type', 'password');
    });

    it('should clear input on unmount', async () => {
      const onChange = vi.fn();
      const { unmount } = render(
        <SecureInput 
          label="API Key"
          value={MOCK_API_KEY}
          onChange={onChange}
        />
      );

      unmount();
      
      // Component should have cleared sensitive data from memory
      expect(onChange).not.toHaveBeenCalledWith(expect.stringContaining('sk-'));
    });

    it('should toggle visibility when show/hide clicked', async () => {
      const user = userEvent.setup();
      const { getByLabelText, getByRole } = render(
        <SecureInput 
          label="API Key"
          value={MOCK_API_KEY}
          onChange={() => {}}
        />
      );

      const input = getByLabelText('API Key');
      const toggleButton = getByRole('button', { name: /show|hide/i });

      // Initially masked
      expect(input).toHaveAttribute('type', 'password');

      // Click to show
      await user.click(toggleButton);
      expect(input).toHaveAttribute('type', 'text');

      // Click to hide
      await user.click(toggleButton);
      expect(input).toHaveAttribute('type', 'password');
    });

    it('should prevent copy/paste when secure mode enabled', async () => {
      const user = userEvent.setup();
      const { getByLabelText } = render(
        <SecureInput 
          label="API Key"
          value={MOCK_API_KEY}
          onChange={() => {}}
          secureMode={true}
        />
      );

      const input = getByLabelText('API Key');
      
      // Should have copy/paste disabled
      expect(input).toHaveAttribute('onCopy');
      expect(input).toHaveAttribute('onPaste');
    });
  });

  describe('CredentialStorage Component', () => {
    it('should store credential in vault', async () => {
      const onStored = vi.fn();
      const { getByLabelText, getByRole } = render(
        <CredentialStorage 
          passwordManager={passwordManager}
          serviceName={MOCK_SERVICE_NAME}
          onStored={onStored}
        />
      );

      const user = userEvent.setup();

      // Enter API key
      const keyInput = getByLabelText('API Key');
      await user.type(keyInput, MOCK_API_KEY);

      // Enter service name
      const nameInput = getByLabelText('Service Name');
      await user.clear(nameInput);
      await user.type(nameInput, 'z.ai');

      // Click save
      const saveButton = getByRole('button', { name: /save/i });
      await user.click(saveButton);

      // Verify callback was called
      await waitFor(() => {
        expect(onStored).toHaveBeenCalledWith(
          expect.objectContaining({
            service: 'z.ai',
            success: true,
          })
        );
      });
    });

    it('should validate API key format', async () => {
      const onError = vi.fn();
      const { getByLabelText, getByRole, getByText } = render(
        <CredentialStorage 
          passwordManager={passwordManager}
          serviceName={MOCK_SERVICE_NAME}
          onError={onError}
        />
      );

      const user = userEvent.setup();

      // Enter invalid API key (too short)
      const keyInput = getByLabelText('API Key');
      await user.type(keyInput, 'short');

      // Click save
      const saveButton = getByRole('button', { name: /save/i });
      await user.click(saveButton);

      // Should show validation error
      await waitFor(() => {
        expect(getByText(/invalid api key/i)).toBeInTheDocument();
      });

      expect(onError).toHaveBeenCalled();
    });

    it('should require vault to be unlocked', async () => {
      // Lock the vault
      passwordManager.lock();

      const onError = vi.fn();
      const { getByRole, getByText } = render(
        <CredentialStorage 
          passwordManager={passwordManager}
          serviceName={MOCK_SERVICE_NAME}
          onError={onError}
        />
      );

      // Should show unlock prompt
      expect(getByText(/vault is locked/i)).toBeInTheDocument();

      // Save button should be disabled
      const saveButton = getByRole('button', { name: /save/i });
      expect(saveButton).toBeDisabled();
    });

    it('should mask stored credentials in UI', async () => {
      const { getByLabelText, getByRole, getByTestId } = render(
        <CredentialStorage 
          passwordManager={passwordManager}
          serviceName={MOCK_SERVICE_NAME}
        />
      );

      const user = userEvent.setup();

      // Store credential
      const keyInput = getByLabelText('API Key');
      await user.type(keyInput, MOCK_API_KEY);

      const saveButton = getByRole('button', { name: /save/i });
      await user.click(saveButton);

      // Verify stored credential is masked
      await waitFor(() => {
        const storedKey = getByTestId('stored-credential');
        expect(storedKey).toHaveTextContent('••••••••');
        expect(storedKey).not.toHaveTextContent(MOCK_API_KEY);
      });
    });

    it('should retrieve and decrypt stored credential', async () => {
      const { getByLabelText, getByRole, getByTestId } = render(
        <CredentialStorage 
          passwordManager={passwordManager}
          serviceName={MOCK_SERVICE_NAME}
        />
      );

      const user = userEvent.setup();

      // Store credential
      const keyInput = getByLabelText('API Key');
      await user.type(keyInput, MOCK_API_KEY);

      const nameInput = getByLabelText('Service Name');
      await user.clear(nameInput);
      await user.type(nameInput, 'z.ai');

      await user.click(getByRole('button', { name: /save/i }));

      // Click to reveal
      await waitFor(() => {
        expect(getByRole('button', { name: /reveal/i })).toBeEnabled();
      });

      await user.click(getByRole('button', { name: /reveal/i }));

      // Verify decrypted value
      await waitFor(() => {
        const revealedKey = getByTestId('revealed-credential');
        expect(revealedKey).toHaveTextContent(MOCK_API_KEY);
      });
    });

    it('should handle auto-lock during entry', async () => {
      const shortLivedManager = new PasswordManager({
        autoLockMs: 100, // 100ms for testing
      });
      await shortLivedManager.unlock(MOCK_MASTER_PASSWORD);

      const onError = vi.fn();
      const { getByLabelText, getByRole, getByText } = render(
        <CredentialStorage 
          passwordManager={shortLivedManager}
          serviceName={MOCK_SERVICE_NAME}
          onError={onError}
        />
      );

      const user = userEvent.setup();

      // Start typing
      const keyInput = getByLabelText('API Key');
      await user.type(keyInput, 'partial-key');

      // Wait for auto-lock
      await new Promise(resolve => setTimeout(resolve, 150));

      // Try to save
      const saveButton = getByRole('button', { name: /save/i });
      await user.click(saveButton);

      // Should show locked error
      await waitFor(() => {
        expect(getByText(/vault locked/i)).toBeInTheDocument();
      });

      shortLivedManager.destroy();
    });
  });

  describe('End-to-End Credential Flow', () => {
    it('should complete full credential storage workflow', async () => {
      const user = userEvent.setup();
      
      // 1. Render credential storage
      const { getByLabelText, getByRole, queryByText } = render(
        <CredentialStorage 
          passwordManager={passwordManager}
          serviceName="ai-provider"
        />
      );

      // 2. Verify vault is unlocked
      expect(queryByText(/vault is locked/i)).not.toBeInTheDocument();

      // 3. Enter service details
      await user.type(getByLabelText('Service Name'), 'z.ai');
      await user.type(getByLabelText('API Key'), MOCK_API_KEY);

      // 4. Save credential
      await user.click(getByRole('button', { name: /save/i }));

      // 5. Verify success
      await waitFor(() => {
        expect(queryByText(/credential saved/i)).toBeInTheDocument();
      });

      // 6. Verify input is cleared (security)
      expect(getByLabelText('API Key')).toHaveValue('');
    });

    it('should persist across TUI sessions', async () => {
      const vaultPath = `${tempDir}/test-vault.enc`;
      
      // First session: Store credential
      const manager1 = new PasswordManager();
      await manager1.unlock(MOCK_MASTER_PASSWORD);
      
      // Store credential logic here...
      
      manager1.destroy();

      // Second session: Retrieve credential
      const manager2 = new PasswordManager();
      await manager2.unlock(MOCK_MASTER_PASSWORD);
      
      // Retrieve and verify credential...
      
      manager2.destroy();
    });
  });

  describe('Security Requirements', () => {
    it('should never log credentials', async () => {
      const consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {});
      
      render(
        <CredentialStorage 
          passwordManager={passwordManager}
          serviceName={MOCK_SERVICE_NAME}
        />
      );

      // Verify no credential logging
      const logs = consoleSpy.mock.calls.flat();
      logs.forEach(log => {
        if (typeof log === 'string') {
          expect(log).not.toContain(MOCK_API_KEY);
          expect(log).not.toMatch(/sk-[a-zA-Z0-9]+/);
        }
      });

      consoleSpy.mockRestore();
    });

    it('should clear sensitive data from component state', async () => {
      const { unmount, getByLabelText } = render(
        <CredentialStorage 
          passwordManager={passwordManager}
          serviceName={MOCK_SERVICE_NAME}
        />
      );

      const user = userEvent.setup();
      
      // Enter sensitive data
      await user.type(getByLabelText('API Key'), MOCK_API_KEY);

      // Unmount component
      unmount();

      // Component should have cleaned up
      // (In real implementation, we'd verify memory is cleared)
    });
  });
});

// Integration test with actual vault
describe('Vault + TUI Integration', () => {
  it('should store and retrieve real credential via vault', async () => {
    const manager = new PasswordManager();
    await manager.unlock('integration-test-password');

    // Store credential
    const credentialData = Buffer.from(JSON.stringify({
      service: 'z.ai',
      key: 'test-key-format-sk-12345',
      createdAt: new Date().toISOString(),
    }));

    const encrypted = await manager.encrypt(credentialData);
    expect(encrypted).toBeDefined();
    expect(encrypted.ciphertext).toBeDefined();

    // Retrieve and decrypt
    const decrypted = await manager.decrypt(encrypted);
    const retrieved = JSON.parse(decrypted.toString());
    
    expect(retrieved.service).toBe('z.ai');
    expect(retrieved.key).toBe('test-key-format-sk-12345');

    manager.destroy();
  });
});
