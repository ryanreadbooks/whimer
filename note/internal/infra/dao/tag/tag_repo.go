package tag

import (
	"context"
	"fmt"
	"time"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xslice"
	"github.com/ryanreadbooks/whimer/misc/xsql"
)

// Tag 标签持久化对象
type Tag struct {
	Id    int64  `db:"id"`    // primary key
	Name  string `db:"name"`  // tag name
	Ctime int64  `db:"ctime"` // create time
}

// TagRepo 标签数据库仓储 - 纯数据库操作
type TagRepo struct {
	db *xsql.DB
}

func NewTagRepo(db *xsql.DB) *TagRepo {
	return &TagRepo{
		db: db,
	}
}

func (r *TagRepo) Create(ctx context.Context, tag *Tag) (int64, error) {
	const sqlInsert = "INSERT INTO tag(name,ctime) VALUES(?,?)"
	if tag.Ctime == 0 {
		tag.Ctime = time.Now().Unix()
	}

	res, err := r.db.ExecCtx(ctx, sqlInsert, tag.Name, tag.Ctime)
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

func (r *TagRepo) FindByName(ctx context.Context, name string) (*Tag, error) {
	const sqlFind = "SELECT id,name,ctime FROM tag WHERE name=?"
	var tag Tag
	err := r.db.QueryRowCtx(ctx, &tag, sqlFind, name)
	return &tag, xerror.Wrap(xsql.ConvertError(err))
}

func (r *TagRepo) FindById(ctx context.Context, id int64) (*Tag, error) {
	const sqlFind = "SELECT id,name,ctime FROM tag WHERE id=?"
	var tag Tag
	err := r.db.QueryRowCtx(ctx, &tag, sqlFind, id)
	return &tag, xerror.Wrap(xsql.ConvertError(err))
}

func (r *TagRepo) BatchGetById(ctx context.Context, ids []int64) ([]*Tag, error) {
	if len(ids) == 0 {
		return []*Tag{}, nil
	}

	const sql = "SELECT id,name,ctime FROM tag WHERE id IN (%s)"
	var tags []*Tag
	err := r.db.QueryRowsCtx(ctx, &tags, fmt.Sprintf(sql, xslice.JoinInts(ids)))
	return tags, xerror.Wrap(xsql.ConvertError(err))
}
