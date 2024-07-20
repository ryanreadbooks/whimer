package repo

import "github.com/zeromicro/go-zero/core/stores/sqlx"

type Repo struct {
	db sqlx.Session
}

func New() *Repo {
	r := &Repo{}

	return r
}
