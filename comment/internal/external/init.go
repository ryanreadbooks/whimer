package external

import (
	"github.com/ryanreadbooks/whimer/comment/internal/config"
	"github.com/ryanreadbooks/whimer/misc/xhttp/middleware/auth"
)

var (
	auther *auth.Auth
	err    error
)

func Init(c *config.Config) {
	auther, err = auth.New(&auth.Config{Addr: c.External.Grpc.Passport})
	if err != nil || auther == nil {
		panic(err)
	}
}

func GetAuther() *auth.Auth {
	return auther
}
