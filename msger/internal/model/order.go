package model

type Order int32

const (
	OrderUnspecified Order = iota
	OrderAsc
	OrderDesc
)

func (o Order) String() string {
	switch o {
	case OrderAsc:
		return "asc"
	case OrderDesc:
		return "desc"
	}
	return "unspecified"
}

func (o Order) Desc() bool {
	return o == OrderDesc
}

func (o Order) Asc() bool {
	return o == OrderAsc
}

func (o Order) Unspecified() bool {
	return o == OrderUnspecified
}
