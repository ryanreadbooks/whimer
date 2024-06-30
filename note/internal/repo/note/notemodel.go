package note

import (
	"context"
	"errors"
	"fmt"

	msqlx "github.com/ryanreadbooks/whimer/misc/sqlx"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ NoteModel = (*customNoteModel)(nil)

var ErrNotFound = sqlx.ErrNotFound

type (
	// NoteModel is an interface to be customized, add more methods here,
	// and implement the added methods in customNoteModel.
	NoteModel interface {
		noteModel
		noteModelExtra
		withSession(session sqlx.Session) NoteModel
	}

	noteModelExtra interface {
		InsertTx(data *Note, callback msqlx.AfterInsert) msqlx.TransactFunc
		UpdateTx(data *Note) msqlx.TransactFunc
		DeleteTx(id int64) msqlx.TransactFunc
		ListByOwner(ctx context.Context, uid int64) ([]*Note, error)
	}

	customNoteModel struct {
		*defaultNoteModel
	}
)

// NewNoteModel returns a model for the database table.
func NewNoteModel(conn sqlx.SqlConn) NoteModel {
	return &customNoteModel{
		defaultNoteModel: newNoteModel(conn),
	}
}

func (m *customNoteModel) withSession(session sqlx.Session) NoteModel {
	return NewNoteModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customNoteModel) InsertTx(data *Note, callback msqlx.AfterInsert) msqlx.TransactFunc {
	return func(ctx context.Context, s sqlx.Session) error {
		query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?, ?)", m.table, noteRowsExpectAutoSet)
		ret, err := s.ExecCtx(ctx, query, data.Title, data.Desc, data.Privacy, data.Owner, data.CreateAt, data.UpdateAt)
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

func (m *customNoteModel) UpdateTx(data *Note) msqlx.TransactFunc {
	return func(ctx context.Context, s sqlx.Session) error {
		query := fmt.Sprintf("update %s set %s where `id` = ?", m.table, noteRowsWithPlaceHolder)
		_, err := s.ExecCtx(ctx, query, data.Title, data.Desc, data.Privacy, data.Owner, data.CreateAt, data.UpdateAt, data.Id)
		return err
	}
}

func (m *customNoteModel) DeleteTx(id int64) msqlx.TransactFunc {
	return func(ctx context.Context, s sqlx.Session) error {
		query := fmt.Sprintf("delete from %s where `id` = ?", m.table)
		_, err := s.ExecCtx(ctx, query, id)
		return err
	}
}

func (m *customNoteModel) ListByOwner(ctx context.Context, uid int64) ([]*Note, error) {
	query := fmt.Sprintf("select %s from %s where `owner` = ?", noteRows, m.table)
	var res []*Note = make([]*Note, 0)
	err := m.conn.QueryRowsCtx(ctx, &res, query, uid)
	switch err {
	case nil:
		return res, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}
