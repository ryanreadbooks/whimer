package vault

import (
	"context"
	"os"
	"testing"
)

func TestEncrypt(t *testing.T) {
	v := New(Option{
		Schema:  "http",
		Host:    "localhost:8200",
		KeyName: "devtest",
		Token:   os.Getenv("ENV_VAULT_ROOT_TOKEN"),
		Context: "hello world",
	})

	ctx := context.TODO()
	enc, err := v.Encrypt(ctx, "this is vault encrypt testing")
	t.Log(err)
	t.Log(enc)

	// decrypt
	plain, err := v.Decrypt(ctx, enc)
	t.Log(err)
	t.Log(plain)
}

func TestEncrypt_NoContext(t *testing.T) {
	v := New(Option{
		Schema:  "http",
		Host:    "localhost:8200",
		KeyName: "nocontext",
		Token:   os.Getenv("ENV_VAULT_ROOT_TOKEN"),
	})

	ctx := context.TODO()
	enc, err := v.Encrypt(ctx, "this is vault encrypt testing")
	t.Log(err)
	t.Log(enc)

	// decrypt
	plain, err := v.Decrypt(ctx, enc)
	t.Log(err)
	t.Log(plain)
}
