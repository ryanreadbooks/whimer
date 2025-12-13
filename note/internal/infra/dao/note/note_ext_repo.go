package note

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xslice"
	"github.com/ryanreadbooks/whimer/misc/xsql"
)

type ExtPO struct {
	NoteId  int64           `db:"note_id"`  // note id
	Tags    string          `db:"tags"`     // note tags: shape like, tag_id1,tag_id2,...,tag_idN
	AtUsers json.RawMessage `db:"at_users"` // at users: json object string: [{"nickname":"user1","uid":1001},{"nickname":"user2","uid":1002}]
	Ctime   int64           `db:"ctime"`
	Utime   int64           `db:"utime"`
}

const (
	extFields = "note_id,tags,at_users,ctime,utime"
)

// NoteExtRepo 笔记扩展信息数据库仓储 - 纯数据库操作
type NoteExtRepo struct {
	db *xsql.DB
}

func NewNoteExtRepo(db *xsql.DB) *NoteExtRepo {
	return &NoteExtRepo{
		db: db,
	}
}

func (d *NoteExtRepo) Upsert(ctx context.Context, ext *ExtPO) error {
	now := time.Now().Unix()
	if ext.Ctime == 0 {
		ext.Ctime = now
	}
	ext.Utime = now

	const sql = "INSERT INTO note_ext(" + extFields + ") VALUES(?,?,?,?,?) " +
		" ON DUPLICATE KEY UPDATE tags=VALUES(tags),at_users=VALUES(at_users),utime=VALUES(utime)"

	_, err := d.db.ExecCtx(ctx, sql, ext.NoteId, ext.Tags, ext.AtUsers, ext.Ctime, ext.Utime)
	return xerror.Wrap(xsql.ConvertError(err))
}

func (d *NoteExtRepo) Delete(ctx context.Context, noteId int64) error {
	const sql = "DELETE FROM note_ext WHERE note_id=? LIMIT 1"
	_, err := d.db.ExecCtx(ctx, sql, noteId)
	return xerror.Wrap(xsql.ConvertError(err))
}

func (d *NoteExtRepo) GetById(ctx context.Context, noteId int64) (*ExtPO, error) {
	const sql = "SELECT " + extFields + " FROM note_ext WHERE note_id=?"
	var ext ExtPO
	err := d.db.QueryRowCtx(ctx, &ext, sql)
	return &ext, xerror.Wrap(xsql.ConvertError(err))
}

func (d *NoteExtRepo) BatchGetById(ctx context.Context, noteIds []int64) ([]*ExtPO, error) {
	var exts []*ExtPO
	const sql = "SELECT " + extFields + " FROM note_ext WHERE note_id IN (%s)"
	err := d.db.QueryRowsCtx(ctx, &exts, fmt.Sprintf(sql, xslice.JoinInts(noteIds)))
	return exts, xerror.Wrap(xsql.ConvertError(err))
}

