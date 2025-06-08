package biz

type Biz struct {
	SessionBiz
}

func New() Biz {
	return Biz{
		SessionBiz: NewSessionBiz(),
	}
}
