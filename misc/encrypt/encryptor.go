package encrypt

import "context"

// 加解密
type Encryptor interface {
	Encrypt(ctx context.Context, plain string) (cipher string, err error)
	Decrypt(ctx context.Context, cipher string) (plain string, err error)
}
