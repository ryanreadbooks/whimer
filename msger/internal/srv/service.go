package srv

import (
	"github.com/ryanreadbooks/whimer/msger/internal/biz"
	"github.com/ryanreadbooks/whimer/msger/internal/config"
	"github.com/ryanreadbooks/whimer/msger/internal/infra"
	"github.com/ryanreadbooks/whimer/msger/internal/infra/dep"
	"github.com/ryanreadbooks/whimer/msger/internal/srv/systemchat"
	"github.com/ryanreadbooks/whimer/msger/internal/srv/userchat"
)

type Service struct {
	SystemChatSrv *systemchat.SystemChatSrv
	UserChatSrv   *userchat.UserChatSrv
}

// 初始化一个service
func NewService(c *config.Config) *Service {
	s := &Service{}
	// 基础设施初始化
	infra.Init(c)
	dep.Init(c)
	biz := biz.New()

	s.SystemChatSrv = systemchat.NewSystemChatSrv(biz)
	s.UserChatSrv = userchat.NewUserChatSrv(biz)

	return s
}
