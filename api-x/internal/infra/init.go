package infra

import "github.com/ryanreadbooks/whimer/api-x/internal/config"

func Init(c *config.Config) {
	InitPassport(c)
	InitNote(c)
	InitCommenter(c)
	InitRelation(c)
	InitMsger(c)
}
