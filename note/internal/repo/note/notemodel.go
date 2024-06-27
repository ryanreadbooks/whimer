package note

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ NoteModel = (*customNoteModel)(nil)

var ErrNotFound = sqlx.ErrNotFound

type (
	// NoteModel is an interface to be customized, add more methods here,
	// and implement the added methods in customNoteModel.
	NoteModel interface {
		noteModel
		withSession(session sqlx.Session) NoteModel
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
