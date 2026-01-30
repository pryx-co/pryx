package vault

import (
	"fmt"
	"sync"
)

// Scope represents the permission level for vault operations
type Scope string

const (
	// ScopeReadOnly allows reading credentials only
	ScopeReadOnly Scope = "readonly"
	// ScopeReadWrite allows reading and writing credentials
	ScopeReadWrite Scope = "readwrite"
	// ScopeAdmin allows all operations including deleting credentials and managing scopes
	ScopeAdmin Scope = "admin"
)

// ScopeManager manages access control scopes for vault operations
type ScopeManager struct {
	mu sync.RWMutex

	// DefaultScope is the default scope for all credentials
	defaultScope Scope

	// CredentialScopes maps credential IDs to their specific scopes
	credentialScopes map[string]Scope

	// AllowedActions defines which actions are permitted for each scope
	allowedActions map[Scope]map[Action]bool
}

// Action represents a vault operation
type Action string

const (
	ActionVaultUnlock Action = "vault:unlock"
	ActionVaultLock   Action = "vault:lock"
	ActionVaultRotate Action = "vault:rotate"

	ActionCredRead   Action = "credential:read"
	ActionCredWrite  Action = "credential:write"
	ActionCredDelete Action = "credential:delete"
	ActionCredList   Action = "credential:list"

	ActionScopeGet    Action = "scope:get"
	ActionScopeSet    Action = "scope:set"
	ActionScopeDelete Action = "scope:delete"
)

// NewScopeManager creates a new scope manager with default settings
func NewScopeManager() *ScopeManager {
	sm := &ScopeManager{
		defaultScope:     ScopeReadWrite,
		credentialScopes: make(map[string]Scope),
		allowedActions:   make(map[Scope]map[Action]bool),
	}

	// Initialize allowed actions for each scope
	sm.allowedActions[ScopeReadOnly] = map[Action]bool{
		ActionVaultUnlock: true,
		ActionVaultLock:   true,
		ActionCredRead:    true,
		ActionCredList:    true,
		ActionScopeGet:    true,
	}

	sm.allowedActions[ScopeReadWrite] = map[Action]bool{
		ActionVaultUnlock: true,
		ActionVaultLock:   true,
		ActionVaultRotate: true,
		ActionCredRead:    true,
		ActionCredWrite:   true,
		ActionCredList:    true,
		ActionScopeGet:    true,
		ActionScopeSet:    true,
	}

	sm.allowedActions[ScopeAdmin] = map[Action]bool{
		ActionVaultUnlock: true,
		ActionVaultLock:   true,
		ActionVaultRotate: true,
		ActionCredRead:    true,
		ActionCredWrite:   true,
		ActionCredDelete:  true,
		ActionCredList:    true,
		ActionScopeGet:    true,
		ActionScopeSet:    true,
		ActionScopeDelete: true,
	}

	return sm
}

// SetDefaultScope sets the default scope for all credentials
func (sm *ScopeManager) SetDefaultScope(scope Scope) error {
	if !sm.isValidScope(scope) {
		return fmt.Errorf("invalid scope: %s", scope)
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.defaultScope = scope
	return nil
}

// GetDefaultScope returns the default scope
func (sm *ScopeManager) GetDefaultScope() Scope {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return sm.defaultScope
}

// SetCredentialScope sets a specific scope for a credential
func (sm *ScopeManager) SetCredentialScope(credentialID string, scope Scope) error {
	if !sm.isValidScope(scope) {
		return fmt.Errorf("invalid scope: %s", scope)
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.credentialScopes[credentialID] = scope
	return nil
}

// GetCredentialScope returns the scope for a specific credential
// Falls back to default scope if not explicitly set
func (sm *ScopeManager) GetCredentialScope(credentialID string) Scope {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if scope, ok := sm.credentialScopes[credentialID]; ok {
		return scope
	}

	return sm.defaultScope
}

// DeleteCredentialScope removes a credential-specific scope
// After deletion, the credential will use the default scope
func (sm *ScopeManager) DeleteCredentialScope(credentialID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	delete(sm.credentialScopes, credentialID)
}

// CanPerform checks if an action is allowed for a given scope
func (sm *ScopeManager) CanPerform(scope Scope, action Action) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if actions, ok := sm.allowedActions[scope]; ok {
		return actions[action]
	}

	return false
}

// CanPerformOnCredential checks if an action is allowed on a specific credential
func (sm *ScopeManager) CanPerformOnCredential(credentialID string, action Action) bool {
	scope := sm.GetCredentialScope(credentialID)
	return sm.CanPerform(scope, action)
}

// CheckPermission returns an error if the action is not allowed
func (sm *ScopeManager) CheckPermission(scope Scope, action Action) error {
	if !sm.CanPerform(scope, action) {
		return fmt.Errorf("permission denied: scope %s cannot perform %s", scope, action)
	}
	return nil
}

// CheckPermissionOnCredential returns an error if the action is not allowed on the credential
func (sm *ScopeManager) CheckPermissionOnCredential(credentialID string, action Action) error {
	scope := sm.GetCredentialScope(credentialID)
	return sm.CheckPermission(scope, action)
}

// GetAllCredentialScopes returns a copy of all credential-specific scopes
func (sm *ScopeManager) GetAllCredentialScopes() map[string]Scope {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	result := make(map[string]Scope, len(sm.credentialScopes))
	for k, v := range sm.credentialScopes {
		result[k] = v
	}

	return result
}

// GetAllowedActions returns all actions allowed for a given scope
func (sm *ScopeManager) GetAllowedActions(scope Scope) []Action {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var actions []Action
	if allowed, ok := sm.allowedActions[scope]; ok {
		for action := range allowed {
			actions = append(actions, action)
		}
	}

	return actions
}

// GetAllScopes returns all available scope levels
func (sm *ScopeManager) GetAllScopes() []Scope {
	return []Scope{ScopeReadOnly, ScopeReadWrite, ScopeAdmin}
}

// isValidScope checks if a scope is valid
func (sm *ScopeManager) isValidScope(scope Scope) bool {
	for _, s := range sm.GetAllScopes() {
		if s == scope {
			return true
		}
	}
	return false
}

// ScopeInfo provides human-readable information about a scope
type ScopeInfo struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Actions     []string `json:"actions"`
}

// GetScopeInfo returns information about a scope
func (sm *ScopeManager) GetScopeInfo(scope Scope) ScopeInfo {
	info := ScopeInfo{
		Name:    string(scope),
		Actions: sm.getActionStrings(sm.GetAllowedActions(scope)),
	}

	switch scope {
	case ScopeReadOnly:
		info.Description = "Read credentials and list vault contents only"
	case ScopeReadWrite:
		info.Description = "Read, write, and rotate credentials"
	case ScopeAdmin:
		info.Description = "Full access including delete and scope management"
	}

	return info
}

// getActionStrings converts Action slices to string slices
func (sm *ScopeManager) getActionStrings(actions []Action) []string {
	result := make([]string, len(actions))
	for i, a := range actions {
		result[i] = string(a)
	}
	return result
}
