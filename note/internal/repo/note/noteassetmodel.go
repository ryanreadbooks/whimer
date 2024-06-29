package note

import (
	"context"
	"errors"
	"fmt"

	msqlx "github.com/ryanreadbooks/whimer/misc/sqlx"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ NoteAssetModel = (*customNoteAssetModel)(nil)

type (
	// NoteAssetModel is an interface to be customized, add more methods here,
	// and implement the added methods in customNoteAssetModel.
	NoteAssetModel interface {
		noteAssetModel
		noteAssetModelTx
		withSession(session sqlx.Session) NoteAssetModel
	}

	customNoteAssetModel struct {
		*defaultNoteAssetModel
	}

	noteAssetModelTx interface {
		InsertTx(data *NoteAsset, callback msqlx.AfterInsert) msqlx.TransactFunc
		UpdateTx(data *NoteAsset) msqlx.TransactFunc
		DeleteTx(id int64) msqlx.TransactFunc
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
		ret, err := s.ExecCtx(ctx, query, data.AssertKey, data.AssertType, data.NoteId, data.CreateAt)
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
		_, err := s.ExecCtx(ctx, query, data.AssertKey, data.AssertType, data.NoteId, data.CreateAt, data.Id)
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
