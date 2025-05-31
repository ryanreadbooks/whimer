package p2p

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ryanreadbooks/whimer/misc/utils/slices"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	gm "github.com/ryanreadbooks/whimer/msger/internal/global/model"
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

func (d *InboxDao) BatchCreate(ctx context.Context, msgs []*InboxMsgPO) error {
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

			args = append(args, data.UserId, data.ChatId, data.MsgId, data.MsgSeq, data.Status, data.Ctime)
		}
		_, err := d.db.ExecCtx(ctx, sql, args...)

		return xsql.ConvertError(err)
	})

	return err
}

// 列出某个用户在某个会话下的消息id
func (d *InboxDao) ListMsg(ctx context.Context,
	userId, chatId, seq int64, cnt int32) ([]int64, error) {

	var msgIds []int64
	const sql = "SELECT msg_id FROM p2p_inbox WHERE user_id=? AND " +
		"chat_id=? AND msg_seq<? ORDER BY msg_seq DESC LIMIT ?"
	err := d.db.QueryRowsCtx(ctx, &msgIds, sql, userId, chatId, seq, cnt)
	return msgIds, xsql.ConvertError(err)
}

// 更新状态为已读（不包含撤回）
func (d *InboxDao) UpdateStatusToRead(ctx context.Context, userId, chatId int64) error {
	sql := "UPDATE p2p_inbox SET status=? WHERE user_id=? AND chat_id=? AND status!=?"
	_, err := d.db.ExecCtx(ctx, sql, gm.InboxRead, userId, chatId, gm.InboxRevoked)
	return xsql.ConvertError(err)
}

// 撤回消息
func (d *InboxDao) RevokeMsg(ctx context.Context, chatId, msgId int64) error {
	sql := "UPDATE p2p_inbox SET status=? WHERE chat_id=? AND msg_id=?"
	_, err := d.db.ExecCtx(ctx, sql, gm.InboxRevoked, chatId, msgId)
	return xsql.ConvertError(err)
}
