package dao

import (
	"github.com/ryanreadbooks/whimer/misc/xsql"
)

const (
	taskPOTableName = "conductor_task"
)

var (
	taskPOFields = xsql.GetFieldSlice(&TaskPO{})
)

type TaskPO struct {
	Id            []byte `db:"id"              json:"id"`
	Namespace     string `db:"namespace"       json:"namespace"`
	TaskType      string `db:"task_type"       json:"task_type"`
	InputArgs     []byte `db:"input_args"      json:"input_args"`
	OutputArgs    []byte `db:"ouput_args"      json:"ouput_args"`
	CallbackUrl   string `db:"callback_url"    json:"callback_url"`
	State         string `db:"state"           json:"state"`
	MaxRetryCnt   int    `db:"max_retry_cnt"   json:"max_retry_cnt"`
	MaxTimeoutSec int    `db:"max_timeout_sec" json:"max_timeout_sec"`
	Utime         int64  `db:"utime"           json:"utime"`
	Version       int64  `db:"version"         json:"version"`
}

func (TaskPO) TableName() string {
	return taskPOTableName
}

func (s *TaskPO) Values() []any {
	inputArgs := s.InputArgs
	if s.InputArgs == nil {
		inputArgs = []byte{}
	}
	outputArgs := s.OutputArgs
	if s.OutputArgs == nil {
		outputArgs = []byte{}
	}
	return []any{
		s.Id,
		s.Namespace,
		s.TaskType,
		inputArgs,
		outputArgs,
		s.CallbackUrl,
		s.State,
		s.MaxRetryCnt,
		s.MaxTimeoutSec,
		s.Utime,
		s.Version,
	}
}
