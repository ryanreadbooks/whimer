package systemnotify

import (
	"context"
	"encoding/json"

	mentionvo "github.com/ryanreadbooks/whimer/pilot/internal/domain/common/mention/vo"
	notifyentity "github.com/ryanreadbooks/whimer/pilot/internal/domain/systemnotify/entity"
	notifyvo "github.com/ryanreadbooks/whimer/pilot/internal/domain/systemnotify/vo"

	"github.com/ryanreadbooks/whimer/misc/uuid"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	noteid "github.com/ryanreadbooks/whimer/note/pkg/id"
)

type mentionMsgContent struct {
	*notifyvo.NotifyAtUsersOnNoteParamContent    `json:"note_content,omitempty"`
	*notifyvo.NotifyAtUsersOnCommentParamContent `json:"comment_content,omitempty"`
	Receivers                                    []*mentionvo.AtUser        `json:"receivers"`
	Loc                                          notifyvo.NotifyMsgLocation `json:"loc"`
}

func (m *mentionMsgContent) GetLocation() notifyvo.NotifyMsgLocation {
	if m.NotifyAtUsersOnNoteParamContent != nil {
		return notifyvo.NotifyMsgOnNote
	}
	return notifyvo.NotifyMsgOnComment
}

func (m *mentionMsgContent) GetSourceUid() int64 {
	if m.NotifyAtUsersOnNoteParamContent != nil {
		return m.NotifyAtUsersOnNoteParamContent.SourceUid
	}
	if m.NotifyAtUsersOnCommentParamContent != nil {
		return m.NotifyAtUsersOnCommentParamContent.SourceUid
	}
	return 0
}

func (m *mentionMsgContent) GetNoteId() noteid.NoteId {
	if m.NotifyAtUsersOnNoteParamContent != nil {
		return m.NotifyAtUsersOnNoteParamContent.NoteId
	}
	if m.NotifyAtUsersOnCommentParamContent != nil {
		return m.NotifyAtUsersOnCommentParamContent.NoteId
	}

	return 0
}

func (m *mentionMsgContent) GetContent() string {
	if m.NotifyAtUsersOnNoteParamContent != nil {
		return m.NotifyAtUsersOnNoteParamContent.NoteDesc
	}
	if m.NotifyAtUsersOnCommentParamContent != nil {
		return m.NotifyAtUsersOnCommentParamContent.Comment
	}
	return ""
}

func (m *mentionMsgContent) GetCommentId() int64 {
	if m.NotifyAtUsersOnCommentParamContent != nil {
		return m.NotifyAtUsersOnCommentParamContent.CommentId
	}
	return 0
}

func parseMentionMsgs(ctx context.Context, rawMsgs []*notifyvo.RawSystemMsg) []*notifyentity.MentionedMsg {
	msgs := make([]*notifyentity.MentionedMsg, 0, len(rawMsgs))

	for _, msg := range rawMsgs {
		mgid, err := uuid.ParseString(msg.Id)
		if err != nil {
			xlog.Msg("parse mention msg id failed").Err(err).Extras("msgid", msg.Id).Errorx(ctx)
			continue
		}

		mm := notifyentity.MentionedMsg{
			Id:     msg.Id,
			SendAt: mgid.UnixSec(),
		}

		if msg.Status != notifyvo.MsgStatusRecalled {
			var v mentionMsgContent
			if err := json.Unmarshal(msg.Content, &v); err != nil {
				xlog.Msg("unmarshal mention msg content failed").Err(err).Errorx(ctx)
				continue
			}

			mm.Type = v.GetLocation()
			mm.Uid = v.GetSourceUid()
			mm.RecvUsers = v.Receivers
			mm.NoteId = v.GetNoteId()
			mm.CommentId = v.GetCommentId()
			mm.Content = v.GetContent()
			mm.Status = notifyvo.MsgStatusNormal
		} else {
			mm.Status = notifyvo.MsgStatusRecalled
		}
		msgs = append(msgs, &mm)
	}

	return msgs
}
