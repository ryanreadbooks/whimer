package external

import (
	"github.com/ryanreadbooks/whimer/note/internal/config"
	"github.com/ryanreadbooks/whimer/note/internal/external/passport"
)

func Init(c *config.Config) {
	err := passport.New(c)
	if err != nil {
		panic(err)
	}
}
