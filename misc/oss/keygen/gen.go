package keygen

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"strconv"
	"strings"
	"time"

	googleuuid "github.com/google/uuid"
	"github.com/ryanreadbooks/whimer/misc/utils"
	"github.com/ryanreadbooks/whimer/misc/uuid"
)

type Stringer interface {
	GetRandomString() string
}

type RandomStringer struct{}

func (s RandomStringer) GetRandomString() string {
	return googleuuid.NewString() + strconv.FormatInt(time.Now().UnixNano(), 10)
}

type RandomStringerV7 struct{}

func (s RandomStringerV7) GetRandomString() string {
	return uuid.NewUUID().String()
}

type Generator struct {
	bucket        string
	prefix        string
	suffix        string
	prependBucket bool
	prependPrefix bool

	stringer Stringer
}

type Option func(*Generator)

func WithBucket(s string) Option {
	return func(g *Generator) {
		g.bucket = s
	}
}

func WithPrefix(p string) Option {
	return func(g *Generator) {
		g.prefix = p
	}
}

func WithSuffix(s string) Option {
	return func(g *Generator) {
		g.suffix = s
	}
}

func WithPrependBucket(p bool) Option {
	return func(g *Generator) {
		g.prependBucket = p
	}
}

func WithPrependPrefix(p bool) Option {
	return func(g *Generator) {
		g.prependPrefix = p
	}
}

func WithStringer(stringer Stringer) Option {
	return func(g *Generator) {
		g.stringer = stringer
	}
}

func NewGenerator(opts ...Option) *Generator {
	gen := &Generator{
		stringer:      RandomStringer{},
		prependBucket: true,
	}

	for _, o := range opts {
		o(gen)
	}

	return gen
}

func (g *Generator) Gen() string {
	// format:
	// /[bucket]/[prefix]/string_[suffix]
	str := g.stringer.GetRandomString()

	var builder strings.Builder
	if len(g.bucket) != 0 {
		builder.WriteByte('/')
		builder.WriteString(g.bucket)
	}

	if len(g.prefix) != 0 {
		builder.WriteByte('/')
		builder.WriteString(g.prefix)
	}

	builder.WriteString(str)
	if len(g.suffix) != 0 {
		builder.WriteByte('_')
		builder.WriteString(g.suffix)
	}

	raw := builder.String()
	b64 := base64.StdEncoding.EncodeToString(utils.StringToBytes(raw))

	hasher := sha1.New()
	hasher.Write(utils.StringToBytes(b64))

	res := hex.EncodeToString(hasher.Sum(nil))
	var prefix string
	if g.prependBucket {
		prefix = g.bucket + "/"
	}

	if g.prependPrefix {
		prefix = prefix + g.prefix + "/"
	}

	return prefix + res
}

// format:
// /[bucket]/[prefix]/string_[suffix]
//
// unwrap bucket and key
func (g *Generator) Unwrap(s string) (bucket, key string, ok bool) {
	if !g.prependBucket {
		return g.bucket, s, true
	}

	// bucket is prepended
	if !strings.HasPrefix(s, g.bucket+"/") {
		ok = false
		return
	}

	key = strings.TrimPrefix(s, g.bucket+"/")
	return g.bucket, key, true
}

// 检查s是否是由该生成器生成的
func (g *Generator) Check(s string) (ok bool) {
	if g.prependBucket {
		if !strings.HasPrefix(s, g.bucket+"/") {
			return false
		}
	}

	s = strings.TrimPrefix(s, g.bucket+"/")

	if g.prependPrefix {
		if !strings.HasPrefix(s, g.prefix+"/") {
			return false
		}
	}

	return true
}

// 去除s中的bucket和prefix前缀 只返回key本身
func (g *Generator) TrimBucketAndPrefix(s string) string {
	if g.prependBucket {
		s = strings.TrimPrefix(s, g.bucket+"/")
	}

	if g.prependPrefix {
		s = strings.TrimPrefix(s, g.prefix+"/")
	}

	return s
}
