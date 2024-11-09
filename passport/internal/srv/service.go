package srv

import (

	"github.com/ryanreadbooks/whimer/passport/internal/biz"
	"github.com/ryanreadbooks/whimer/passport/internal/config"
	"github.com/ryanreadbooks/whimer/passport/internal/infra"
)

type Service struct {
	c *config.Config

	// domain services
	UserSrv   *UserSrv
	AccessSrv *AccessSrv
}

func New(c *config.Config) *Service {
	s := &Service{
		c: c,
	}

	infra.Init(c)
	biz := biz.New()
	s.UserSrv = NewUserSrv(s, biz)
	s.AccessSrv = NewAccessSrv(s, biz)

	return s
}
