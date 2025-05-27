package encrypt

import "context"

// 加解密
type Encryptor interface {
	Encrypt(ctx context.Context, plain string) (enc string, err error)
	Decrypt(ctx context.Context, enc string) (plain string, err error)
}
