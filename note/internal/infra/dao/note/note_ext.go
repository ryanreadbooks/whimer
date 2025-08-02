package note

import (
	"context"
	"time"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xsql"
)

type Ext struct {
	NoteId int64  `db:"note_id"` // note id
	Tags   string `db:"tags"`    // note tags: shape like, tag_id1,tag_id2,...,tag_idN
	Ctime  int64  `db:"ctime"`
	Utime  int64  `db:"utime"`
}

type NoteExtDao struct {
	db *xsql.DB
}

func NewNoteExtDao(db *xsql.DB) *NoteExtDao {
	return &NoteExtDao{
		db: db,
	}
}

func (d *NoteExtDao) Upsert(ctx context.Context, ext *Ext) error {
	now := time.Now().Unix()
	if ext.Ctime == 0 {
		ext.Ctime = now
	}
	ext.Utime = now

	const sql = "INSERT INTO note_ext(note_id,tags,ctime,utime) VALUES(?,?,?,?) " +
		" ON DUPLICATE KEY UPDATE tags=VALUES(tags),utime=VALUES(utime)"
		
	_, err := d.db.ExecCtx(ctx, sql, ext.NoteId, ext.Tags, ext.Ctime, ext.Utime)
	return xerror.Wrap(xsql.ConvertError(err))
}

func (d *NoteExtDao) Delete(ctx context.Context, noteId int64) error {
	const sql = "DELETE FROM note_ext WHERE note_id=? LIMIT 1"
	_, err := d.db.ExecCtx(ctx, sql, noteId)
	return xerror.Wrap(xsql.ConvertError(err))
}
