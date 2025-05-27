package encrypt

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/ryanreadbooks/whimer/misc/encrypt"
	"github.com/ryanreadbooks/whimer/misc/utils"
)

// Vault Docs: https://developer.hashicorp.com/vault/tutorials/encryption-as-a-service/eaas-transit#encrypt-secrets
type Manager struct {
	opt *VaultOption
	// TODO http client
}

type VaultOption struct {
	Addr string `json:"addr" yaml:"addr"`
}

func NewManager(opt *VaultOption) *Manager {
	return &Manager{
		opt: opt,
	}
}

func (m *Manager) NameEncryptor(name string) encrypt.Encryptor {
	return &namedVault{
		name:    name,
		encPath: fmt.Sprintf("/v1/transit/encrypt/%s", name),
		decPath: fmt.Sprintf("/v1/transit/decrypt/%s", name),
	}
}

type namedVault struct {
	name    string
	encPath string
	decPath string
	// TODO http client
}

func (nv *namedVault) Encrypt(ctx context.Context, plain string) (enc string, err error) {
	// do base64 first
	data := base64.StdEncoding.EncodeToString(utils.StringToBytes(plain))
	body := fmt.Sprintf(`{"plaintext": "%s"}`, data)
	// add http header

	_ = body
	return "", nil
}

func (nv *namedVault) Decrypt(ctx context.Context, enc string) (plain string, err error) {
	return "", nil
}
