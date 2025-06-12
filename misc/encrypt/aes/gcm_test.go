package aes

import (
	"context"
	"testing"
)

func Test_AESEncryptor(t *testing.T) {
	encryptor, err := NewAes256GCMEncryptor("abcdefghijklmnop")
	t.Log(err)
	ctx := context.Background()
	c, err := encryptor.Encrypt(ctx, "hello-world")
	t.Log(err)
	t.Log(c)

	plain, err := encryptor.Decrypt(ctx, c)
	t.Log(err)
	t.Log(plain)
}

func Test_AESEncryptor2(t *testing.T) {
	encryptor, err := NewAes256GCMEncryptor("abcdefghijklmnop", WithHex())
	t.Log(err)
	ctx := context.Background()
	c, err := encryptor.Encrypt(ctx, "hello-world")
	t.Log(err)
	t.Log(c)

	plain, err := encryptor.Decrypt(ctx, c)
	t.Log(err)
	t.Log(plain)
}

func Test_AESEncryptorFixNonce(t *testing.T) {
	encryptor, err := NewAes256GCMEncryptor("7a9d41b8c05f3e26a1b9d07c8f6e3a5d",
		WithFixNonce([]byte("2c7a9b3e5d4f1a80")))
	t.Log(err)
	ctx := context.Background()
	c, err := encryptor.Encrypt(ctx, "ello-eowl")
	t.Log(err)
	t.Log(c)

	plain, err := encryptor.Decrypt(ctx, c)
	t.Log(err)
	t.Log(plain)
}
