package biz

import (
	"fmt"

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
	}
}
