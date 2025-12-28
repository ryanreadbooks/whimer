package worker

import (
	"encoding/json"

	taskv1 "github.com/ryanreadbooks/whimer/conductor/api/task/v1"
)

// Task 任务信息
type Task struct {
	Id          string
	Namespace   string
	TaskType    string
	InputArgs   []byte
	CallbackUrl string
	MaxRetryCnt int64
	ExpireTime  int64
	Ctime       int64
	TraceId     string
}

// UnmarshalInput 反序列化输入参数
func (t *Task) UnmarshalInput(v any) error {
	if len(t.InputArgs) == 0 {
		return nil
	}
	return json.Unmarshal(t.InputArgs, v)
}

func taskFromProto(t *taskv1.Task) *Task {
	if t == nil {
		return nil
	}
	return &Task{
		Id:          t.Id,
		Namespace:   t.Namespace,
		TaskType:    t.TaskType,
		InputArgs:   t.InputArgs,
		CallbackUrl: t.CallbackUrl,
		MaxRetryCnt: t.MaxRetryCnt,
		ExpireTime:  t.ExpireTime,
		Ctime:       t.Ctime,
		TraceId:     t.TraceId,
	}
}
