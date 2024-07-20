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
	auther, err = auth.New(&auth.Config{c.ThreeRd.Grpc.Passport})
	if err != nil || auther == nil {
		panic(err)
	}
}

func GetAuther() *auth.Auth {
	return auther
}
