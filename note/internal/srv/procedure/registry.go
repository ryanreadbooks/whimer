package procedure

import "github.com/ryanreadbooks/whimer/note/internal/model"

type Registry struct {
	procedures map[model.ProcedureType]Procedure
}

func NewRegistry() *Registry {
	return &Registry{
		procedures: make(map[model.ProcedureType]Procedure),
	}
}

func (r *Registry) Register(p Procedure) {
	r.procedures[p.Type()] = p
}

func (r *Registry) Get(protype model.ProcedureType) (Procedure, bool) {
	p, ok := r.procedures[protype]
	return p, ok
}

// 获取所有已注册的流程类型
func (r *Registry) Types() []model.ProcedureType {
	types := make([]model.ProcedureType, 0, len(r.procedures))
	for t := range r.procedures {
		types = append(types, t)
	}
	return types
}
