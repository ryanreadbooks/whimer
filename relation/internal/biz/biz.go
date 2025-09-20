package biz

type Biz struct {
	Relation           *RelationBiz
	RelationSettingBiz *RelationSettingBiz
}

func New() Biz {
	return Biz{
		Relation:           NewRelationBiz(),
		RelationSettingBiz: NewRelationSettingBiz(),
	}
}
