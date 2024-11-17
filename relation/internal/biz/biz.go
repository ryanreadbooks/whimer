package biz

type Biz struct {
	Relation RelationBiz
}

func New() Biz {
	return Biz{
		Relation: NewRelationBiz(),
	}
}
