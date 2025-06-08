package vault

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/ryanreadbooks/whimer/misc/encrypt"
	"github.com/ryanreadbooks/whimer/misc/utils"
	xclient "github.com/ryanreadbooks/whimer/misc/xhttp/client"
)

const (
	vaultReqTokenHeaderKey = "X-Vault-Token"
)

type vaultResponse struct {
	RequestId     string            `json:"request_id"`
	LeaseId       string            `json:"lease_id"`
	Renewable     bool              `json:"renewable"`
	LeaseDuration int               `json:"lease_duration"`
	Data          vaultResponseData `json:"data"`
}

type vaultResponseData struct {
	Ciphertext string `json:"ciphertext,omitempty"`
	Plaintext  string `json:"plaintext,omitempty"`
	KeyVersion int64  `json:"key_version,omitempty"`
}

type Option struct {
	Schema  string `json:"schema" yaml:"schema"`
	Host    string `json:"host" yaml:"host"`
	KeyName string `json:"key_name" yaml:"key_name"`
	Token   string `json:"token" yaml:"token"`
	// 指定了context(base64格式)后，相同的明文会加密成相同的密文, 但是创建token需要支持
	Context string `json:"context" yaml:"context"`
}

// Vault Docs: https://developer.hashicorp.com/vault/tutorials/encryption-as-a-service/eaas-transit#encrypt-secrets
type Valut struct {
	opt *Option
	cli *xclient.Client

	encPath, decPath string
}

func New(opt Option) encrypt.Encryptor {
	if opt.Context != "" {
		_, err := base64.StdEncoding.DecodeString(opt.Context)
		if err != nil {
			opt.Context = base64.StdEncoding.EncodeToString(utils.StringToBytes(opt.Context))
		}
	}

	v := &Valut{
		opt:     &opt,
		encPath: fmt.Sprintf("/v1/transit/encrypt/%s", opt.KeyName),
		decPath: fmt.Sprintf("/v1/transit/decrypt/%s", opt.KeyName),
	}

	v.cli = xclient.New(opt.Schema, opt.Host)

	return v
}

func (v *Valut) Encrypt(ctx context.Context, plain string) (string, error) {
	// do base64 first
	data := base64.StdEncoding.EncodeToString(utils.StringToBytes(plain))
	var reqBody string
	if v.opt.Context != "" {
		reqBody = fmt.Sprintf(`{"plaintext": "%s", "context": "%s"}`, data, v.opt.Context)
	} else {
		reqBody = fmt.Sprintf(`{"plaintext": "%s"}`, data)
	}

	req, err := http.NewRequestWithContext(ctx,
		http.MethodPost,
		v.encPath,
		strings.NewReader(reqBody),
	)
	if err != nil {
		return "", fmt.Errorf("vault failed to create request: %w", err)
	}
	req.Header.Add(vaultReqTokenHeaderKey, v.opt.Token)

	var respData vaultResponse
	_, err = v.cli.Fetch(req, &respData)
	if err != nil {
		return "", fmt.Errorf("vault failed to encrypt: %w", err)
	}

	return respData.Data.Ciphertext, nil
}

func (v *Valut) Decrypt(ctx context.Context, ciphertext string) (string, error) {
	var reqBody string
	if v.opt.Context != "" {
		reqBody = fmt.Sprintf(`{"ciphertext": "%s", "context": "%s"}`, ciphertext, v.opt.Context)
	} else {
		reqBody = fmt.Sprintf(`{"ciphertext": "%s"}`, ciphertext)
	}

	req, err := http.NewRequestWithContext(ctx,
		http.MethodPost,
		v.decPath,
		strings.NewReader(reqBody),
	)
	if err != nil {
		return "", fmt.Errorf("vault failed to create request: %w", err)
	}
	req.Header.Add(vaultReqTokenHeaderKey, v.opt.Token)

	var respData vaultResponse
	_, err = v.cli.Fetch(req, &respData)
	if err != nil {
		return "", fmt.Errorf("vault failed to decrypt: %w", err)
	}

	plaintext, err := base64.StdEncoding.DecodeString(respData.Data.Plaintext)
	if err != nil {
		return "", fmt.Errorf("vault failed to base64 decode: %w", err)
	}

	return string(plaintext), nil
}
