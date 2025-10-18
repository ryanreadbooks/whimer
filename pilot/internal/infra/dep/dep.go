package dep

import "github.com/ryanreadbooks/whimer/pilot/internal/config"

func Init(c *config.Config) {
	InitPassport(c)
	InitNote(c)
	InitCommenter(c)
	InitRelation(c)
	InitMsger(c)
	InitSearch(c)
	InitWsLink(c)
}
