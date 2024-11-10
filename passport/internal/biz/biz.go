package biz

type Biz struct {
	Access    AccessBiz
	AccessSms AccessSmsBiz
	User      UserBiz
	Register  RegisterBiz
}

func New() Biz {
	return Biz{
		Access:    NewAccessBiz(),
		AccessSms: NewAccessSmsBiz(),
		User:      NewUserBiz(),
		Register:  NewRegisterBiz(),
	}
}
