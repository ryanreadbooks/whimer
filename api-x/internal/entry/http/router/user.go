package router

import (
	"github.com/ryanreadbooks/whimer/api-x/internal/entry/http/handler"
	"github.com/ryanreadbooks/whimer/api-x/internal/entry/http/middleware"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
)

// 用户信息相关路由
func regUserRoutes(group *xhttp.RouterGroup, h *handler.Handler) {
	g := group.Group("/user", middleware.CanLogin())
	{
		v1g := g.Group("/v1")
		{
			v1gMustLogin := v1g.Group("", middleware.MustLoginCheck())
			{
				// 批量拉取用户信息
				v1gMustLogin.Get("/info/list", h.User.ListInfos())
				// 获取用户粉丝列表
				v1gMustLogin.Get("/fans", h.User.ListUserFans())
				// 获取用户关注列表
				v1gMustLogin.Get("/followings", h.User.ListUserFollowings())

				// 用户设置相关
				{
					settingsGroup := v1gMustLogin.Group("/settings")
					{
						// 获取全部设置
						settingsGroup.Get("/all", h.User.GetAllSettings())
						// 设置粉丝列表/关注列表展示情况
						settingsGroup.Post("/relation/update", h.Relation.UpdateSettings())
					}
				}
			}

			// 拉取单个用户的信息
			v1g.Get("/get", h.User.GetUser())

			// 获取用户的投稿数量、点赞数量等信息
			v1g.Get("/stat", h.User.GetUserStat())

			// 用户hover卡片信息
			v1g.Get("/hover/profile", h.User.GetHoverProfile())
		}
	}
}
