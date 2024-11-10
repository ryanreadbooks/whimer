package csrf

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/ryanreadbooks/whimer/misc/utils"
)

const tokenSizeHalf = 16
const tokenSize = tokenSizeHalf * 2

type Token string

func (t Token) Cookie(domain string, expire time.Time) *http.Cookie {
	return &http.Cookie{
		Name:    cookieCsrfName,
		Value:   string(t),
		Domain:  domain,
		Path:    "/",
		Expires: expire,
	}
}

func GetToken() Token {
	token := randomBytes()

	return Token(maskToken(token))
}

func randomBytes() []byte {
	buf := make([]byte, tokenSizeHalf)
	_, err := rand.Read(buf)
	if err != nil {
		buf = utils.RandomByte(tokenSizeHalf)
	}

	return buf
}

func maskToken(token []byte) (s string) {
	mask := randomBytes()
	return hex.EncodeToString(xor(token, mask))
}

func xor(a, b []byte) []byte {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}

	out := make([]byte, n)

	for i := 0; i < n; i++ {
		out[i] = a[i] ^ b[i]
	}

	return out
}
