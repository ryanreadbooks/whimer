package note

import (
	"context"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xslice"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/note/internal/model"
)

const (
	noteTableName = "note"
)

var (
	noteFields    = xsql.GetFieldSlice(&NotePO{})
	noteInsFields = xsql.GetFieldSlice(&NotePO{}, "id") // 插入时不包含 id
)

// NotePO 笔记持久化对象
type NotePO struct {
	Id       int64           `db:"id"`
	Title    string          `db:"title"`   // 标题
	Desc     string          `db:"desc"`    // 描述
	Privacy  model.Privacy   `db:"privacy"` // 公开类型
	Owner    int64           `db:"owner"`   // 笔记作者
	Ip       []byte          `db:"ip"`
	NoteType model.NoteType  `db:"note_type"` // 笔记类型
	State    model.NoteState `db:"state"`     // 状态
	CreateAt int64           `db:"create_at"` // 笔记创建时间
	UpdateAt int64           `db:"update_at"` // 笔记更新时间
	ModifyAt int64           `db:"modify_at"` // 数据记录更新时间, 和笔记更新时间update_at区分开
}

func (NotePO) TableName() string {
	return noteTableName
}

func (n *NotePO) Values() []any {
	return []any{
		n.Id,
		n.Title,
		n.Desc,
		n.Privacy,
		n.Owner,
		n.Ip,
		n.NoteType,
		n.State,
		n.CreateAt,
		n.UpdateAt,
		n.ModifyAt,
	}
}

func (n *NotePO) InsertValues() []any {
	var ip []byte = n.Ip
	if n.Ip == nil {
		ip = []byte{}
	}

	return []any{
		n.Title,
		n.Desc,
		n.Privacy,
		n.Owner,
		ip,
		n.NoteType,
		n.State,
		n.CreateAt,
		n.UpdateAt,
		n.ModifyAt,
	}
}

// NoteRepo 笔记数据库仓储 - 纯数据库操作
type NoteRepo struct {
	db *xsql.DB
}

func NewNoteRepo(db *xsql.DB) *NoteRepo {
	return &NoteRepo{
		db: db,
	}
}

func (r *NoteRepo) GetNoteType(ctx context.Context, id int64) (model.NoteType, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select("note_type").From(noteTableName).Where(sb.Equal("id", id))
	sql, args := sb.Build()
	var noteType model.NoteType
	err := r.db.QueryRowCtx(ctx, &noteType, sql, args...)
	return noteType, xerror.Wrap(xsql.ConvertError(err))
}

func (r *NoteRepo) FindOne(ctx context.Context, id int64) (*NotePO, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(noteFields...)
	sb.From(noteTableName)
	sb.Where(sb.Equal("id", id))
	sql, args := sb.Build()
	resp := new(NotePO)
	err := r.db.QueryRowCtx(ctx, resp, sql, args...)
	return resp, xerror.Wrap(xsql.ConvertError(err))
}

func (r *NoteRepo) FindOneForUpdate(ctx context.Context, id int64) (*NotePO, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(noteFields...)
	sb.From(noteTableName)
	sb.Where(sb.Equal("id", id))
	sb.ForUpdate()

	sql, args := sb.Build()
	resp := new(NotePO)
	err := r.db.QueryRowCtx(ctx, resp, sql, args...)
	if err != nil {
		return nil, xerror.Wrap(xsql.ConvertError(err))
	}

	return resp, nil
}

// BatchGet 批量获取笔记
func (r *NoteRepo) BatchGet(ctx context.Context, ids []int64) ([]*NotePO, error) {
	if len(ids) == 0 {
		return []*NotePO{}, nil
	}

	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(noteFields...)
	sb.From(noteTableName)
	sb.Where(sb.In("id", xslice.Any(ids)...))

	sql, args := sb.Build()
	var notes []*NotePO
	err := r.db.QueryRowsCtx(ctx, &notes, sql, args...)
	if err != nil {
		return nil, xerror.Wrap(xsql.ConvertError(err))
	}

	return notes, nil
}

// 必须带条件
func (r *NoteRepo) List(ctx context.Context, conds ...NoteRepoCondition) ([]*NotePO, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(noteFields...)
	sb.From(noteTableName)

	nrc := &noteRepoCondition{
		sb: sb,
	}
	for _, cond := range conds {
		cond(nrc)
	}

	sql, args := sb.Build()
	res := make([]*NotePO, 0)
	err := r.db.QueryRowsCtx(ctx, &res, sql, args...)
	if err != nil {
		return nil, xerror.Wrap(xsql.ConvertError(err))
	}
	return res, nil
}

func (r *NoteRepo) ListByCursor(
	ctx context.Context, cursor int64, limit int32,
	conds ...NoteRepoCondition,
) ([]*NotePO, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(noteFields...)
	sb.From(noteTableName)
	sb.Where(sb.LessThan("id", cursor))

	nrc := &noteRepoCondition{
		sb: sb,
	}
	for _, cond := range conds {
		cond(nrc)
	}

	sb.OrderByDesc("create_at")
	sb.OrderByDesc("id")
	sb.Limit(int(limit))

	sql, args := sb.Build()
	res := make([]*NotePO, 0)
	err := r.db.QueryRowsCtx(ctx, &res, sql, args...)
	if err != nil {
		return nil, xerror.Wrap(xsql.ConvertError(err))
	}
	return res, nil
}

// ListByPage 分页查询笔记（支持 offset/limit 方式）
func (r *NoteRepo) ListByPage(
	ctx context.Context, page, count int32,
	conds ...NoteRepoCondition,
) ([]*NotePO, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(noteFields...)
	sb.From(noteTableName)

	nrc := &noteRepoCondition{
		sb: sb,
	}
	for _, cond := range conds {
		cond(nrc)
	}

	sb.OrderByDesc("create_at")
	sb.OrderByDesc("id")
	sb.Offset(int((page - 1) * count))
	sb.Limit(int(count))

	sql, args := sb.Build()
	res := make([]*NotePO, 0, count)
	err := r.db.QueryRowsCtx(ctx, &res, sql, args...)
	if err != nil {
		return nil, xerror.Wrap(xsql.ConvertError(err))
	}
	return res, nil
}

// Count 统计笔记数量（支持 conditions）
func (r *NoteRepo) Count(ctx context.Context, conds ...NoteRepoCondition) (int64, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select("COUNT(*)")
	sb.From(noteTableName)

	nrc := &noteRepoCondition{
		sb: sb,
	}
	for _, cond := range conds {
		cond(nrc)
	}

	sql, args := sb.Build()
	var cnt int64
	err := r.db.QueryRowCtx(ctx, &cnt, sql, args...)
	return cnt, xerror.Wrap(xsql.ConvertError(err))
}

func (r *NoteRepo) Insert(ctx context.Context, note *NotePO) (int64, error) {
	ib := sqlbuilder.NewInsertBuilder()
	ib.InsertInto(noteTableName)
	ib.Cols(noteInsFields...)
	ib.Values(note.InsertValues()...)

	sql, args := ib.Build()
	res, err := r.db.ExecCtx(ctx, sql, args...)
	if err != nil {
		return 0, xerror.Wrap(xsql.ConvertError(err))
	}
	newId, err := res.LastInsertId()
	if err != nil {
		return 0, xerror.Wrap(xsql.ConvertError(err))
	}

	return int64(newId), nil
}

// 更新笔记
func (r *NoteRepo) Update(ctx context.Context, note *NotePO) error {
	ub := sqlbuilder.NewUpdateBuilder()
	ub.Update(noteTableName)
	ub.Set(
		ub.EQ("title", note.Title),
		ub.EQ("`desc`", note.Desc),
		ub.EQ("privacy", note.Privacy),
		ub.EQ("owner", note.Owner),
		ub.EQ("ip", note.Ip),
		ub.EQ("note_type", note.NoteType),
		ub.EQ("state", note.State),
		ub.EQ("update_at", time.Now().Unix()),
		ub.EQ("modify_at", time.Now().Unix()),
	)
	ub.Where(ub.EQ("id", note.Id))

	sql, args := ub.Build()
	_, err := r.db.ExecCtx(ctx, sql, args...)

	return xerror.Wrap(xsql.ConvertError(err))
}

// GetOwner 获取笔记的owner
func (r *NoteRepo) GetOwner(ctx context.Context, noteId int64) (int64, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select("owner").From(noteTableName).Where(sb.EQ("id", noteId))
	sql, args := sb.Build()

	var ownerId int64
	err := r.db.QueryRowCtx(ctx, &ownerId, sql, args...)
	if err != nil {
		return 0, xerror.Wrap(xsql.ConvertError(err))
	}

	return ownerId, nil
}

func (r *NoteRepo) Delete(ctx context.Context, noteId int64) error {
	db := sqlbuilder.NewDeleteBuilder()
	db.DeleteFrom(noteTableName).Where(db.EQ("id", noteId))
	sql, args := db.Build()
	_, err := r.db.ExecCtx(ctx, sql, args...)

	return xerror.Wrap(xsql.ConvertError(err))
}

func (r *NoteRepo) GetPublicByCursor(ctx context.Context, id int64, count int) ([]*NotePO, error) {
	return r.getByCursor(ctx, id, count, model.PrivacyPublic)
}

func (r *NoteRepo) GetPrivateByCursor(ctx context.Context, id int64, count int) ([]*NotePO, error) {
	return r.getByCursor(ctx, id, count, model.PrivacyPrivate)
}

func (r *NoteRepo) getByCursor(ctx context.Context, id int64, count int, privacy model.Privacy) ([]*NotePO, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(noteFields...).
		From(noteTableName).
		Where(sb.GreaterEqualThan("id", id), sb.EQ("privacy", privacy))
	if privacy == model.PrivacyPublic {
		sb.Where(sb.EQ("state", model.NoteStatePublished))
	}

	sb.Limit(count)

	sql, args := sb.Build()
	res := make([]*NotePO, 0, count)
	err := r.db.QueryRowsCtx(ctx, &res, sql, args...)
	return res, xerror.Wrap(xsql.ConvertError(err))
}

func (r *NoteRepo) GetPublicLastId(ctx context.Context) (int64, error) {
	return r.getLastId(ctx, model.PrivacyPublic)
}

func (r *NoteRepo) GetPrivateLastId(ctx context.Context) (int64, error) {
	return r.getLastId(ctx, model.PrivacyPrivate)
}

func (r *NoteRepo) getLastId(ctx context.Context, privacy model.Privacy) (int64, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select("id").
		From(noteTableName).
		Where(sb.EQ("privacy", privacy))
	if privacy == model.PrivacyPublic {
		sb.Where(sb.EQ("state", model.NoteStatePublished))
	}
	sb.OrderByDesc("id").Limit(1)

	sql, args := sb.Build()
	var lastId int64
	err := r.db.QueryRowCtx(ctx, &lastId, sql, args...)
	return lastId, xerror.Wrap(xsql.ConvertError(err))
}

func (r *NoteRepo) UpdateState(ctx context.Context, noteId int64, state model.NoteState) error {
	ub := sqlbuilder.NewUpdateBuilder()
	ub.Update(noteTableName)
	ub.Set(ub.EQ("state", state), ub.EQ("modify_at", time.Now().Unix()))
	ub.Where(ub.EQ("id", noteId))

	sql, args := ub.Build()
	_, err := r.db.ExecCtx(ctx, sql, args...)

	return xerror.Wrap(xsql.ConvertError(err))
}

// 状态升级
func (r *NoteRepo) UpgradeState(ctx context.Context, noteId int64, state model.NoteState) error {
	ub := sqlbuilder.NewUpdateBuilder()
	ub.Update(noteTableName)
	ub.Set(ub.EQ("state", state), ub.EQ("modify_at", time.Now().Unix()))
	ub.Where(ub.EQ("id", noteId), ub.LTE("state", state)) // 当前状态小于等于目标状态才更新

	sql, args := ub.Build()
	_, err := r.db.ExecCtx(ctx, sql, args...)
	return xerror.Wrap(xsql.ConvertError(err))
}
