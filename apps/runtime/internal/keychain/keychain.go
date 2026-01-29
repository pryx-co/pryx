package keychain

import (
	"fmt"
	"strings"

	"github.com/zalando/go-keyring"
)

type Keychain struct {
	service string
}

func New(service string) *Keychain {
	return &Keychain{service: service}
}

func (k *Keychain) Set(user, password string) error {
	return keyring.Set(k.service, user, password)
}

func (k *Keychain) Get(user string) (string, error) {
	return keyring.Get(k.service, user)
}

func (k *Keychain) Delete(user string) error {
	return keyring.Delete(k.service, user)
}

func (k *Keychain) SetProviderKey(provider, key string) error {
	keyName := fmt.Sprintf("provider:%s", provider)
	return k.Set(keyName, key)
}

func (k *Keychain) GetProviderKey(provider string) (string, error) {
	keyName := fmt.Sprintf("provider:%s", provider)
	return k.Get(keyName)
}

func (k *Keychain) DeleteProviderKey(provider string) error {
	keyName := fmt.Sprintf("provider:%s", provider)
	return k.Delete(keyName)
}

func (k *Keychain) ListProviderKeys() ([]string, error) {
	return []string{}, nil
}

func (k *Keychain) MigrateConfigKey(provider, key string) error {
	if key == "" {
		return nil
	}
	return k.SetProviderKey(provider, key)
}

func GetKeyForProvider(provider string) string {
	return fmt.Sprintf("provider:%s", provider)
}

func ExtractProviderFromKey(key string) (string, bool) {
	if strings.HasPrefix(key, "provider:") {
		return strings.TrimPrefix(key, "provider:"), true
	}
	return "", false
}
