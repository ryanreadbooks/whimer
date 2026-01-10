package biz

import (
	"fmt"
	"time"

	notedao "github.com/ryanreadbooks/whimer/note/internal/infra/dao/note"
	"github.com/ryanreadbooks/whimer/note/internal/model"
)

type ProcedureRecord struct {
	Id            int64
	NoteId        int64
	Protype       model.ProcedureType
	TaskId        string
	Status        model.ProcedureStatus
	Ctime         int64
	Utime         int64
	CurRetry      int
	MaxRetryCnt   int
	NextCheckTime int64
	Params        []byte
	ExpiredTime   int64 // 任务过期时间，unix second
}

func (r *ProcedureRecord) GetLockKey() string {
	return fmt.Sprintf("note:procedure:record:lock:%d", r.Id)
}

func ProcedureRecordFromPO(po *notedao.ProcedureRecordPO) *ProcedureRecord {
	return &ProcedureRecord{
		Id:            po.Id,
		NoteId:        po.NoteId,
		Protype:       po.Protype,
		TaskId:        po.TaskId,
		Status:        po.Status,
		Ctime:         po.Ctime,
		Utime:         po.Utime,
		CurRetry:      po.CurRetry,
		MaxRetryCnt:   po.MaxRetryCnt,
		NextCheckTime: po.NextCheckTime,
		Params:        po.Params,
		ExpiredTime:   po.ExpiredTime,
	}
}

// IsExpired 检查任务是否已过期
func (r *ProcedureRecord) IsExpired() bool {
	if r.ExpiredTime <= 0 {
		return false // 兼容旧数据，未设置过期时间视为未过期
	}
	return time.Now().Unix() >= r.ExpiredTime
}
