package model

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	pbmsg "github.com/ryanreadbooks/whimer/msger/api/msg"
	pbuserchatv1 "github.com/ryanreadbooks/whimer/msger/api/userchat/v1"
	uservo "github.com/ryanreadbooks/whimer/pilot/internal/domain/user/vo"
	"github.com/ryanreadbooks/whimer/pilot/internal/model/errors"
)

type MsgReq struct {
	Type    MsgType     `json:"type"`
	Cid     string      `json:"cid"`
	Content *MsgContent `json:"content"`
}

func (m *MsgReq) SetContentType() {
	m.Content.contentType = m.Type
}

func (m *MsgReq) Validate(_ context.Context) error {
	if m == nil {
		return xerror.ErrNilArg
	}

	if !IsValidMsgType(m.Type) {
		return errors.ErrUnsupportedMsgType
	}

	// check content
	if err := m.Content.Validate(); err != nil {
		return err
	}

	return nil
}

// assign msg content as pb format for pbMsg
func AssignPbMsgReqContent(msg *MsgReq, pbMsg *pbuserchatv1.MsgReq) error {
	switch msg.Type {
	case MsgText:
		pbMsg.Content = msg.Content.Text.AsReqPb()
		return nil
	}

	return errors.ErrUnsupportedMsgType
}

// Msg model definition
type Msg struct {
	Id        string       `json:"id,omitempty"`
	Cid       string       `json:"cid,omitempty"`
	Type      MsgType      `json:"type,omitempty"`
	Status    MsgStatus    `json:"status,omitempty"`
	Mtime     int64        `json:"mtime,omitempty"`
	SenderUid int64        `json:"sender_uid,omitempty"`
	Sender    *uservo.User `json:"sender,omitempty"`
	Content   *MsgContent  `json:"content,omitempty"`
	Pos       int64        `json:"pos"`
	Ext       *MsgExt      `json:"ext,omitempty"`
}

func MsgFromChatMsgPb(pbChatMsg *pbuserchatv1.ChatMsg) *Msg {
	if pbChatMsg == nil {
		return &Msg{Type: MsgTypeUnspecified}
	}

	pbMsg := pbChatMsg.GetMsg()

	msg := &Msg{
		Id:        pbMsg.GetId(),
		Type:      MsgType(pbMsg.GetType()),
		Cid:       pbMsg.GetCid(),
		Status:    MsgStatus(pbChatMsg.GetMsg().GetStatus()),
		Mtime:     pbMsg.GetMtime(),
		SenderUid: pbMsg.GetSender(),
		Pos:       pbChatMsg.GetPos(),
		Ext:       MsgExtFromPb(pbMsg.GetExt()),
	}

	// assign content
	if msg.Id != "" && msg.Status != MsgStatusRecall {
		msg.Content = &MsgContent{contentType: msg.Type}
		msg.FillMsgContent(pbMsg)
	}

	return msg
}

func (m *Msg) SetSenderFromVo(u *uservo.User) {
	m.Sender = u
}

func (m *Msg) FillMsgContent(pb *pbmsg.Msg) {
	switch m.Content.contentType {
	case MsgText:
		m.Content.Text = &MsgTextContent{
			Content: pb.GetText().GetContent(),
			Preview: pb.GetText().GetPreview(),
		}
	case MsgImage:
	}
}

type MsgExt struct {
	Recall *MsgExtRecall `json:"recall,omitempty"`
}

type MsgExtRecall struct {
	RecallUid int64 `json:"recall_uid"`
	RecallAt  int64 `json:"recall_at"`
}

func MsgExtFromPb(pbext *pbmsg.MsgExt) *MsgExt {
	if pbext == nil {
		return nil
	}

	ext := &MsgExt{}
	if pbext.Recall != nil {
		ext.Recall = &MsgExtRecall{
			RecallUid: pbext.Recall.GetUid(),
			RecallAt:  pbext.Recall.GetTime(),
		}
	}

	return ext
}
