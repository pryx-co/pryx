package keychain

import (
	"os"
	"testing"
)

func skipIfNoKeyring(t *testing.T) {
	if os.Getenv("CI") == "true" {
		k := New("test-ci-check")
		err := k.Set("test-user", "test-password")
		if err != nil {
			t.Skip("Keyring service not available in CI environment")
		}
		k.Delete("test-user")
	}
}

func TestNew(t *testing.T) {
	service := "test-service"
	k := New(service)

	if k == nil {
		t.Fatal("Expected non-nil keychain")
	}

	if k.service != service {
		t.Errorf("Expected service '%s', got '%s'", service, k.service)
	}
}

func TestSetGetDelete(t *testing.T) {
	skipIfNoKeyring(t)

	// Use a unique service name for testing to avoid conflicts
	service := "pryx-test-" + t.Name()
	k := New(service)

	testUser := "test-user-" + t.Name()
	testPassword := "test-password-123"

	// Clean up any existing test data
	defer func() {
		k.Delete(testUser) // Ignore errors during cleanup
	}()

	// Test Set operation
	err := k.Set(testUser, testPassword)
	if err != nil {
		t.Fatalf("Failed to set password: %v", err)
	}

	// Test Get operation
	password, err := k.Get(testUser)
	if err != nil {
		t.Fatalf("Failed to get password: %v", err)
	}

	if password != testPassword {
		t.Errorf("Expected password '%s', got '%s'", testPassword, password)
	}

	// Test Delete operation
	err = k.Delete(testUser)
	if err != nil {
		t.Fatalf("Failed to delete password: %v", err)
	}

	// Verify deletion
	_, err = k.Get(testUser)
	if err == nil {
		t.Error("Expected error when getting deleted password")
	}
}

func TestGetNonExistent(t *testing.T) {
	service := "pryx-test-nonexistent"
	k := New(service)

	_, err := k.Get("nonexistent-user")
	if err == nil {
		t.Error("Expected error when getting non-existent user")
	}
}

func TestDeleteNonExistent(t *testing.T) {
	service := "pryx-test-nonexistent"
	k := New(service)

	// Keychain may return an error if secret not found
	k.Delete("nonexistent-user") // Ignore errors - different backends behave differently
}

func TestMultipleUsers(t *testing.T) {
	skipIfNoKeyring(t)

	service := "pryx-test-multi-" + t.Name()
	k := New(service)

	users := []struct {
		user     string
		password string
	}{
		{"user1", "password1"},
		{"user2", "password2"},
		{"user3", "password3"},
	}

	// Clean up
	defer func() {
		for _, u := range users {
			k.Delete(u.user)
		}
	}()

	// Set passwords for multiple users
	for _, u := range users {
		err := k.Set(u.user, u.password)
		if err != nil {
			t.Fatalf("Failed to set password for %s: %v", u.user, err)
		}
	}

	// Verify each user's password
	for _, u := range users {
		password, err := k.Get(u.user)
		if err != nil {
			t.Fatalf("Failed to get password for %s: %v", u.user, err)
		}

		if password != u.password {
			t.Errorf("Expected password '%s' for user %s, got '%s'", u.password, u.user, password)
		}
	}
}
