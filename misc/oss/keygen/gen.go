package keygen

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ryanreadbooks/whimer/misc/utils"
)

type Stringer interface {
	GetRandomString() string
}

type RandomStringer struct{}

func (s RandomStringer) GetRandomString() string {
	return uuid.NewString() + strconv.FormatInt(time.Now().UnixNano(), 10)
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
		stringer: RandomStringer{},
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
