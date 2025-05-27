package p2p

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ryanreadbooks/whimer/misc/utils/slices"
	"github.com/ryanreadbooks/whimer/misc/xsql"
)

type InboxDao struct {
	db *xsql.DB
}

func NewInboxDao(db *xsql.DB) *InboxDao {
	return &InboxDao{
		db: db,
	}
}

func (d *InboxDao) DB() *xsql.DB {
	return d.db
}

func (d *InboxDao) BatchCreate(ctx context.Context, msgs []*InboxMsg) error {
	if len(msgs) == 0 {
		return nil
	}

	now := time.Now().UnixNano()

	err := slices.BatchExec(msgs, 100, func(start, end int) error {
		datas := msgs[start:end]
		qm := "(" + insInboxQst + ")"
		qsts := strings.Join(slices.Repeat(qm, len(datas)), ",") // (?,?,?),(?,?,?)
		// 批量插入
		sql := fmt.Sprintf("INSERT INTO p2p_inbox(%s) VALUES %s", insInboxFields, qsts)
		args := make([]any, 0, len(datas)*4)
		for _, data := range datas {
			if data.Ctime == 0 {
				data.Ctime = now
			}

			args = append(args, data.UserId, data.ChatId, data.MsgId, data.Status, data.Ctime)
		}
		_, err := d.db.ExecCtx(ctx, sql, args...)

		return xsql.ConvertError(err)
	})

	return err
}
