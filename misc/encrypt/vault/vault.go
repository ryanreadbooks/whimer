package encrypt

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ryanreadbooks/whimer/misc/encrypt"
	"github.com/ryanreadbooks/whimer/misc/utils"
	"github.com/ryanreadbooks/whimer/misc/xhttp/client"
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
	Cipertext  string `json:"cipertext"`
	KeyVersion string `json:"key_version"`
}

type Option struct {
	Addr    string `json:"addr" yaml:"addr"`
	KeyName string `json:"key_name" yaml:"key_name"`
	Token   string `json:"token" yaml:"token"`
	// 指定了context(base64格式)后，相同的明文会加密成相同的密文, 但是创建token需要支持
	Context string `json:"context" yaml:"context"`
}

// Vault Docs: https://developer.hashicorp.com/vault/tutorials/encryption-as-a-service/eaas-transit#encrypt-secrets
type Valut struct {
	opt *Option
	cli *client.Client

	encPath, decPath string
}

func New(opt *Option) encrypt.Encryptor {
	v := &Valut{
		opt:     opt,
		encPath: fmt.Sprintf("/v1/transit/encrypt/%s", opt.KeyName),
		decPath: fmt.Sprintf("/v1/transit/decrypt/%s", opt.KeyName),
	}

	return v
}

func (v *Valut) Encrypt(ctx context.Context, plain string) (enc string, err error) {
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

	resp, err := v.cli.Do(req)
	if err != nil {
		return "", fmt.Errorf("vault failed to encrypt: %w", err)
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("vault failed to read respbody: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("vault return non ok: %s", resp.Status)
	}

	var respData vaultResponse
	err = json.Unmarshal(respBody, &respData)
	if err != nil {
		return "", fmt.Errorf("valut failed to unmarshal body: %w", err)
	}

	return respData.Data.Cipertext, nil
}

func (v *Valut) Decrypt(ctx context.Context, enc string) (plain string, err error) {
	return "", nil
}
