package aes

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"github.com/ryanreadbooks/whimer/misc/encrypt"
	"github.com/ryanreadbooks/whimer/misc/utils"
)

const (
	defaultNonceSize = 16
)

type Aes256GCMEncryptor struct {
	key         []byte
	cipherBlock cipher.Block

	opt   *opt
	ser   func([]byte) string
	deSer func(string) ([]byte, error)
}

type opt struct {
	useBase64   bool
	hasFixNonce bool
	fixNonce    []byte
}

type Option func(*opt)

func WithHex() Option {
	return func(o *opt) {
		o.useBase64 = false
	}
}

// 设置固定的nonce
//
// 设置了固定的nonce后 相同的明文每次加密会得到相同的密文,
// 否则相同的明文每次加密得到的密文都不同
func WithFixNonce(n []byte) Option {
	return func(o *opt) {
		o.fixNonce = n
		o.hasFixNonce = true
	}
}

func NewAes256GCMEncryptor(key string, opts ...Option) (encrypt.Encryptor, error) {
	k := []byte(key)
	b, err := aes.NewCipher(k)
	if err != nil {
		return nil, err
	}

	o := &opt{
		useBase64: true,
	}

	for _, opt := range opts {
		opt(o)
	}

	serFn := base64.StdEncoding.EncodeToString
	deSerFn := base64.StdEncoding.DecodeString
	if !o.useBase64 {
		serFn = hex.EncodeToString
		deSerFn = hex.DecodeString
	}

	e := &Aes256GCMEncryptor{
		key:         []byte(key),
		cipherBlock: b,
		opt:         o,
		ser:         serFn,
		deSer:       deSerFn,
	}

	return e, nil
}

func (e *Aes256GCMEncryptor) Encrypt(ctx context.Context, plain string) (string, error) {
	nonceSize := defaultNonceSize
	if e.opt.hasFixNonce {
		nonceSize = len(e.opt.fixNonce)
	}

	gcm, err := cipher.NewGCMWithNonceSize(e.cipherBlock, nonceSize)
	if err != nil {
		return "", err
	}

	var nonce []byte = make([]byte, nonceSize)
	if !e.opt.hasFixNonce {
		_, err = rand.Read(nonce)
		if err != nil {
			return "", err
		}
	} else {
		nonce = e.opt.fixNonce
	}

	// nonce放在前面nonceSize字节，密文追加在后面
	temp := gcm.Seal(nonce, nonce, utils.StringToBytes(plain), nil)

	return e.ser(temp), nil
}

func (e *Aes256GCMEncryptor) Decrypt(ctx context.Context, ciphertext string) (string, error) {
	temp, err := e.deSer(ciphertext)
	if err != nil {
		return "", err
	}

	nonceSize := defaultNonceSize
	if e.opt.hasFixNonce {
		nonceSize = len(e.opt.fixNonce)
	}

	gcm, err := cipher.NewGCMWithNonceSize(e.cipherBlock, nonceSize)
	if err != nil {
		return "", err
	}

	nc := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext is too short")
	}

	nonce, rawCipherText := temp[:nc], temp[nc:]
	result, err := gcm.Open(nil, nonce, rawCipherText, nil)
	if err != nil {
		return "", err
	}

	return utils.Bytes2String(result), nil
}
