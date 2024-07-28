package external

import (
	"github.com/ryanreadbooks/whimer/misc/xhttp/middleware/auth"
	"github.com/ryanreadbooks/whimer/note/internal/config"
)

var (
	auther *auth.Auth
	err    error
)

func Init(c *config.Config) {
	auther = auth.MustAuther(c.External.Grpc.Passport)
}

func GetAuther() *auth.Auth {
	return auther
}
