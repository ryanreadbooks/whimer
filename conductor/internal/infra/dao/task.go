package dao

import (
	"github.com/ryanreadbooks/whimer/misc/uuid"
	"github.com/ryanreadbooks/whimer/misc/xsql"
)

const (
	taskPOTableName = "conductor_task"
)

var (
	taskPOFields = xsql.GetFieldSlice(&TaskPO{})
)

type TaskPO struct {
	Id            uuid.UUID `db:"id"              json:"id"`
	Namespace     string    `db:"namespace"       json:"namespace"`
	TaskType      string    `db:"task_type"       json:"task_type"`
	TaskTypeShard int       `db:"task_type_shard" json:"task_type_shard"`
	InputArgs     []byte    `db:"input_args"      json:"input_args"`
	OutputArgs    []byte    `db:"output_args"     json:"output_args"`
	CallbackUrl   string    `db:"callback_url"    json:"callback_url"`
	State         string    `db:"state"           json:"state"`
	TraceId       string    `db:"trace_id"        json:"trace_id"`
	Ctime         int64     `db:"ctime"           json:"ctime"`
	Utime         int64     `db:"utime"           json:"utime"`
	MaxRetryCnt   int64     `db:"max_retry_cnt"   json:"max_retry_cnt"` // -1表示无限重试直到超时, 0表示不重试
	ExpireTime    int64     `db:"expire_time"     json:"expire_time"`   // 任务过期时间 unix ms
	Settings      []byte    `db:"settings"        json:"settings"`      // 额外设置
	Version       int64     `db:"version"         json:"version"`
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
	settings := s.Settings
	if s.Settings == nil {
		settings = []byte{}
	}
	return []any{
		s.Id,
		s.Namespace,
		s.TaskType,
		s.TaskTypeShard,
		inputArgs,
		outputArgs,
		s.CallbackUrl,
		s.State,
		s.TraceId,
		s.Ctime,
		s.Utime,
		s.MaxRetryCnt,
		s.ExpireTime,
		settings,
		s.Version,
	}
}
