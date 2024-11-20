package dep

import "github.com/ryanreadbooks/whimer/feed/internal/config"

func Init(c *config.Config) {
	initNote(c)
	initPassport(c)
}
