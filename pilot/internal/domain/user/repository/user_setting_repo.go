package repository

import (
	"context"

	"github.com/ryanreadbooks/whimer/pilot/internal/domain/user/entity"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/user/vo"
)

// 用户设置
type UserSettingRepository interface {
	// 获取本地存储的用户设置
	GetLocalSetting(ctx context.Context, uid int64) (*entity.UserSetting, error)

	// 获取本地存储的用户设置（加锁）
	GetLocalSettingForUpdate(ctx context.Context, uid int64) (*entity.UserSetting, error)

	// 更新或创建本地用户设置
	UpsertLocalSetting(ctx context.Context, setting *entity.UserSetting) error

	// 获取关系服务的用户设置
	GetRelationSetting(ctx context.Context, uid int64) (*vo.RelationSetting, error)

	// 获取完整用户设置（聚合本地+远程）
	GetFullSetting(ctx context.Context, uid int64) (*vo.FullUserSetting, error)
}
