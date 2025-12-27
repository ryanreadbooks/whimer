package note

import (
	"github.com/huandu/go-sqlbuilder"
	"github.com/ryanreadbooks/whimer/misc/xslice"
	"github.com/ryanreadbooks/whimer/note/internal/model"
)

type noteRepoCondition struct {
	sb *sqlbuilder.SelectBuilder
}

type NoteRepoCondition func(*noteRepoCondition)

func WithNoteStateEqual(state model.NoteState) NoteRepoCondition {
	return func(c *noteRepoCondition) {
		c.sb.Where(c.sb.EQ("state", state))
	}
}

func WithNoteStateIn(states ...model.NoteState) NoteRepoCondition {
	return func(c *noteRepoCondition) {
		if len(states) == 0 {
			return
		}

		c.sb.Where(c.sb.In("state", xslice.Any(states)...))
	}
}

func WithNoteStateNotIn(states ...model.NoteState) NoteRepoCondition {
	return func(c *noteRepoCondition) {
		if len(states) == 0 {
			return
		}

		c.sb.Where(c.sb.NotIn("state", xslice.Any(states)...))
	}
}

func WithNoteOwnerEqual(uid int64) NoteRepoCondition {
	return func(c *noteRepoCondition) {
		c.sb.Where(c.sb.EQ("owner", uid))
	}
}

func WithNoteTypeEqual(noteType model.NoteType) NoteRepoCondition {
	return func(c *noteRepoCondition) {
		c.sb.Where(c.sb.EQ("note_type", noteType))
	}
}

func WithNotePrivacyEqual(privacy model.Privacy) NoteRepoCondition {
	return func(c *noteRepoCondition) {
		c.sb.Where(c.sb.EQ("privacy", privacy))
	}
}
