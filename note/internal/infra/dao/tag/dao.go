package tag

import (
	"context"
	"time"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xsql"
)

type TagDao struct {
	db *xsql.DB
}

func NewTagDao(db *xsql.DB) *TagDao {
	return &TagDao{
		db: db,
	}
}

func (d *TagDao) Create(ctx context.Context, tag *Tag) (int64, error) {
	const sqlInsert = "INSERT INTO tag(name,ctime) VALUES(?,?)"
	if tag.Ctime == 0 {
		tag.Ctime = time.Now().Unix()
	}

	res, err := d.db.ExecCtx(ctx, sqlInsert, tag.Name, tag.Ctime)
	if err != nil {
		err = xsql.ConvertError(err)
		return 0, xerror.Wrap(err)
	}

	newId, err := res.LastInsertId()
	if err != nil {
		return 0, xerror.Wrap(xsql.ConvertError(err))
	}

	return newId, nil
}

func (d *TagDao) rawFind(ctx context.Context, name string) (*Tag, error) {
	const sqlFind = "SELECT id,name,ctime FROM tag WHERE name=?"
	var tag Tag
	err := d.db.QueryRowCtx(ctx, &tag, sqlFind, name)
	return &tag, err
}

func (d *TagDao) Find(ctx context.Context, name string) (*Tag, error) {
	t, err := d.rawFind(ctx, name)
	return t, xerror.Wrap(err)
}

func (d *TagDao) FindById(ctx context.Context, id int64) (*Tag, error) {
	const sqlFind = "SELECT id,name,ctime FROM tag WHERE id=?"
	var tag Tag
	err := d.db.QueryRowCtx(ctx, &tag, sqlFind, id)
	return &tag, xerror.Wrap(xsql.ConvertError(err))
}
