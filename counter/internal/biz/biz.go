package biz

type Biz struct {
	CounterBiz *CounterBiz
}

func New() Biz {
	b := Biz{}
	b.CounterBiz = NewCounterBiz()
	return b
}
