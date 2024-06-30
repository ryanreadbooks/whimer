package note

import (
	"context"
	"errors"
	"fmt"
	"strings"

	msqlx "github.com/ryanreadbooks/whimer/misc/sqlx"
	uslices "github.com/ryanreadbooks/whimer/misc/utils/slices"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ NoteAssetModel = (*customNoteAssetModel)(nil)

type (
	// NoteAssetModel is an interface to be customized, add more methods here,
	// and implement the added methods in customNoteAssetModel.
	NoteAssetModel interface {
		noteAssetModel
		noteAssetModelExtra
		withSession(session sqlx.Session) NoteAssetModel
	}

	customNoteAssetModel struct {
		*defaultNoteAssetModel
	}

	noteAssetModelExtra interface {
		InsertTx(data *NoteAsset, callback msqlx.AfterInsert) msqlx.TransactFunc
		UpdateTx(data *NoteAsset) msqlx.TransactFunc
		DeleteTx(id int64) msqlx.TransactFunc
		BatchInsertTx(datas []*NoteAsset) msqlx.TransactFunc
		DeleteByNoteIdTx(noteId int64, exclude []string) msqlx.TransactFunc
		FindByNoteIdTx(ctx context.Context, sess sqlx.Session, noteId int64) ([]*NoteAsset, error)
		FindByNoteIds(ctx context.Context, noteIds []int64) ([]*NoteAsset, error)
	}
)

// NewNoteAssetModel returns a model for the database table.
func NewNoteAssetModel(conn sqlx.SqlConn) NoteAssetModel {
	return &customNoteAssetModel{
		defaultNoteAssetModel: newNoteAssetModel(conn),
	}
}

func (m *customNoteAssetModel) withSession(session sqlx.Session) NoteAssetModel {
	return NewNoteAssetModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customNoteAssetModel) InsertTx(data *NoteAsset, callback msqlx.AfterInsert) msqlx.TransactFunc {
	return func(ctx context.Context, s sqlx.Session) error {
		query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?)", m.table, noteAssetRowsExpectAutoSet)
		ret, err := s.ExecCtx(ctx, query, data.AssetKey, data.AssetType, data.NoteId, data.CreateAt)
		if err != nil {
			return err
		}
		id, _ := ret.LastInsertId()
		cnt, _ := ret.RowsAffected()
		if id <= 0 || cnt <= 0 {
			return errors.New("insert failure")
		}

		if callback != nil {
			callback(id, cnt)
		}

		return nil
	}
}

func (m *customNoteAssetModel) UpdateTx(data *NoteAsset) msqlx.TransactFunc {
	return func(ctx context.Context, s sqlx.Session) error {
		query := fmt.Sprintf("update %s set %s where `id` = ?", m.table, noteAssetRowsWithPlaceHolder)
		_, err := s.ExecCtx(ctx, query, data.AssetKey, data.AssetType, data.NoteId, data.CreateAt, data.Id)
		return err
	}
}

func (m *customNoteAssetModel) DeleteTx(id int64) msqlx.TransactFunc {
	return func(ctx context.Context, s sqlx.Session) error {
		query := fmt.Sprintf("delete from %s where `id` = ?", m.table)
		_, err := s.ExecCtx(ctx, query, id)
		return err
	}
}

// 批量插入
func (m *customNoteAssetModel) BatchInsertTx(datas []*NoteAsset) msqlx.TransactFunc {
	return func(ctx context.Context, s sqlx.Session) error {
		if len(datas) == 0 {
			return nil
		}

		tmpl := "(?, ?, ?, ?)"
		var builder strings.Builder
		var args []any = make([]any, 0, len(datas)*4)
		for i, data := range datas {
			builder.WriteString(tmpl)
			args = append(args, data.AssetKey, data.AssetType, data.NoteId, data.CreateAt)
			if i != len(datas)-1 {
				builder.WriteByte(',')
			}
		}

		// insert into %s (%s) values (?,?,?,?),(?,?,?,?)
		query := fmt.Sprintf("insert into %s (%s) values %s", m.table, noteAssetRowsExpectAutoSet, builder.String())
		_, err := s.ExecCtx(ctx, query, args...)
		return err
	}
}

func (m *customNoteAssetModel) DeleteByNoteIdTx(noteId int64, assetKeys []string) msqlx.TransactFunc {
	return func(ctx context.Context, s sqlx.Session) error {
		query := fmt.Sprintf("delete from %s where `note_id` = ?", m.table)
		var alen = len(assetKeys)
		var args []any = make([]any, 0, alen)
		args = append(args, noteId)

		if alen != 0 {
			var tmpl string
			for i, ask := range assetKeys {
				tmpl += "?"
				args = append(args, ask)
				if i != alen-1 {
					tmpl += ","
				}
			}
			query += fmt.Sprintf(" and `asset_key` not in (%s)", tmpl)
		}
		_, err := s.ExecCtx(ctx, query, args...)
		return err
	}
}

func (m *customNoteAssetModel) FindByNoteIdTx(ctx context.Context, sess sqlx.Session, noteId int64) ([]*NoteAsset, error) {
	var res = make([]*NoteAsset, 0)
	query := fmt.Sprintf("select %s from %s where `note_id` = ?", noteAssetRows, m.table)
	err := sess.QueryRowsCtx(ctx, &res, query, noteId)
	switch err {
	case nil:
		return res, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customNoteAssetModel) FindByNoteIds(ctx context.Context, noteIds []int64) ([]*NoteAsset, error) {
	if len(noteIds) == 0 {
		return []*NoteAsset{}, nil
	}

	var res = make([]*NoteAsset, 0)
	query := fmt.Sprintf("select %s from %s where `note_id` in (%s)", noteAssetRows, m.table, uslices.JoinInts(noteIds))
	err := m.conn.QueryRowsCtx(ctx, &res, query)
	switch err {
	case nil:
		return res, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}
