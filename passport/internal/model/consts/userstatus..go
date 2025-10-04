package consts

type UserStatus int8

const (
	UserStatusUnknown     UserStatus = 0
	UserStatusNormal      UserStatus = 1 // 正常
	UserStatusAtRisk      UserStatus = 2 // 存在风险
	UserStatusBanned      UserStatus = 3 // 封禁
	UserStatusDeactivated UserStatus = 4 // 注销
)

func (s UserStatus) Unknown() bool     { return s == UserStatusUnknown }
func (s UserStatus) Normal() bool      { return s == UserStatusNormal }
func (s UserStatus) AtRisk() bool      { return s == UserStatusAtRisk }
func (s UserStatus) Banned() bool      { return s == UserStatusBanned }
func (s UserStatus) Deactivated() bool { return s == UserStatusDeactivated }
