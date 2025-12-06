package dao

import (
	"github.com/ryanreadbooks/whimer/misc/xsql"
)

const (
	taskHistoryPOTableName = "conductor_task_history"
)

var (
	taskHistoryPOFields = xsql.GetFieldSlice(&TaskHistoryPO{})
)

type TaskHistoryPO struct {
	Id       int64  `db:"id"        json:"id"`
	TaskId   []byte `db:"task_id"   json:"task_id"`
	State    string `db:"state"     json:"state"`
	RetryCnt int    `db:"retry_cnt" json:"retry_cnt"`
	Ctime    int64  `db:"ctime"     json:"ctime"`
}

func (TaskHistoryPO) TableName() string {
	return taskHistoryPOTableName
}

func (s *TaskHistoryPO) Values() []any {
	return []any{
		s.Id,
		s.TaskId,
		s.State,
		s.RetryCnt,
		s.Ctime,
	}
}
