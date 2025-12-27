package data

import (
	notedao "github.com/ryanreadbooks/whimer/note/internal/infra/dao/note"
	"github.com/ryanreadbooks/whimer/note/internal/model"
)

// NoteCondition 笔记查询条件（重新导出给 Biz 层使用）
type NoteCondition = notedao.NoteRepoCondition

// WithNoteOwnerEqual 指定笔记所有者
func WithNoteOwnerEqual(uid int64) NoteCondition {
	return notedao.WithNoteOwnerEqual(uid)
}

// WithNoteStateEqual 指定笔记状态
func WithNoteStateEqual(state model.NoteState) NoteCondition {
	return notedao.WithNoteStateEqual(state)
}

func WithNoteStateIn(states ...model.NoteState) NoteCondition {
	return notedao.WithNoteStateIn(states...)
}

// WithNoteStateNotIn 排除指定状态
func WithNoteStateNotIn(states ...model.NoteState) NoteCondition {
	return notedao.WithNoteStateNotIn(states...)
}

// WithNoteTypeEqual 指定笔记类型
func WithNoteTypeEqual(noteType model.NoteType) NoteCondition {
	return notedao.WithNoteTypeEqual(noteType)
}

// WithNotePrivacyEqual 指定笔记隐私类型
func WithNotePrivacyEqual(privacy model.Privacy) NoteCondition {
	return notedao.WithNotePrivacyEqual(privacy)
}

