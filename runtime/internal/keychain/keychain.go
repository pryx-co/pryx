package keychain

import "github.com/zalando/go-keyring"

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
