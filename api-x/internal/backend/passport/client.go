package passport

import (
	"github.com/ryanreadbooks/whimer/api-x/internal/config"
	"github.com/ryanreadbooks/whimer/misc/xhttp/middleware/auth"
)

var (
	auther *auth.Auth
	err    error
)

func Init(c *config.Config) {
	auther = auth.MustAuther(c.Backend.Passport)
}

func GetAuther() *auth.Auth {
	return auther
}
