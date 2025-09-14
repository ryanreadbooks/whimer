package biz

import (
	"encoding/base64"
	"fmt"
	"strings"

	counterv1 "github.com/ryanreadbooks/whimer/counter/api/v1"
	recorddao "github.com/ryanreadbooks/whimer/counter/internal/infra/dao/record"

	"github.com/ryanreadbooks/whimer/misc/obfuscate"
	"github.com/ryanreadbooks/whimer/misc/xstring"
)

func NewPbRecord(r *recorddao.Record) *counterv1.Record {
	act := counterv1.RecordAct_RECORD_ACT_UNSPECIFIED
	switch r.Act {
	case recorddao.ActDo:
		act = counterv1.RecordAct_RECORD_ACT_ADD
	case recorddao.ActUndo:
		act = counterv1.RecordAct_RECORD_ACT_UNADD
	}
	return &counterv1.Record{
		BizCode: r.BizCode,
		Uid:     r.Uid,
		Oid:     r.Oid,
		Act:     act,
		Ctime:   r.Ctime,
		Mtime:   r.Mtime,
	}
}

type PageListOrder int8

const (
	PageListDescOrder PageListOrder = 0
	PageListAscOrder  PageListOrder = 1
)

type PageListRecordsParam struct {
	Cursor string
	Count  int32
	Order  PageListOrder
}

func (r *PageListRecordsParam) ParseCursor(obs obfuscate.Obfuscate) (mtime, id int64, err error) {
	raw, err := base64.RawStdEncoding.DecodeString(r.Cursor)
	if err != nil {
		return
	}

	s := xstring.FromBytes(raw)
	unpacked := strings.SplitN(s, ":", 2)
	if len(unpacked) != 2 {
		err = fmt.Errorf("%s is invalid cursor", s)
		return
	}

	mtimeStr := unpacked[0]
	mixIdStr := unpacked[1]
	mtime, err = obs.DeMix(mtimeStr)
	if err != nil {
		err = fmt.Errorf("invalid mtime: %w", err)
		return
	}

	id, err = obs.DeMix(mixIdStr)
	if err != nil {
		err = fmt.Errorf("invalid id: %w", err)
	}

	return
}

func (PageListRecordsParam) FormatCursor(mtime, id int64, obs obfuscate.Obfuscate) string {
	mtimeMix, _ := obs.Mix(mtime)
	idMix, _ := obs.Mix(id)
	cursor := mtimeMix + ":" + idMix
	return base64.RawStdEncoding.EncodeToString(xstring.AsBytes(cursor))
}

type PageResult struct {
	NextCursor string
	HasNext    bool
}

func pbRecordFromDaoRecord(data *recorddao.Record) *counterv1.Record {
	return &counterv1.Record{
		BizCode: int32(data.BizCode),
		Uid:     data.Uid,
		Oid:     data.Oid,
		Act:     counterv1.RecordAct(data.Act),
		Ctime:   data.Ctime,
		Mtime:   data.Mtime,
	}
}
