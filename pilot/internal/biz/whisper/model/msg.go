package model

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	pbmsg "github.com/ryanreadbooks/whimer/msger/api/msg"
	userchatv1 "github.com/ryanreadbooks/whimer/msger/api/userchat/v1"
	usermodel "github.com/ryanreadbooks/whimer/pilot/internal/biz/user/model"
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
func AssignPbMsgReqContent(msg *MsgReq, pbMsg *userchatv1.MsgReq) error {
	switch msg.Type {
	case MsgText:
		pbMsg.Content = msg.Content.Text.AsReqPb()
		return nil
	}

	return errors.ErrUnsupportedMsgType
}

// Msg model definition

type Msg struct {
	Id        string          `json:"id,omitempty"`
	Cid       string          `json:"cid,omitempty"`
	Type      MsgType         `json:"type,omitempty"`
	Status    MsgStatus       `json:"status,omitempty"`
	Mtime     int64           `json:"mtime,omitempty"`
	SenderUid int64           `json:"sender_uid,omitempty"`
	Sender    *usermodel.User `json:"sender,omitempty"`
	Content   *MsgContent     `json:"content,omitempty"`
	// TODO Ext
}

func MsgFromPb(pbm *pbmsg.Msg) *Msg {
	if pbm == nil {
		return &Msg{Type: MsgTypeUnspecified}
	}

	msg := &Msg{
		Id:        pbm.Id,
		Type:      MsgType(pbm.Type),
		Cid:       pbm.Cid,
		Status:    MsgStatus(pbm.Status),
		Mtime:     pbm.Mtime,
		SenderUid: pbm.Sender,
	}

	// assign content
	if msg.Id != "" {
		msg.Content = &MsgContent{contentType: msg.Type}
		msg.FillMsgContent(pbm)
	}

	return msg
}

func (m *Msg) SetSender(u *usermodel.User) {
	m.Sender = u
}

func (m *Msg) FillMsgContent(pb *pbmsg.Msg) {
	// TODO 处理status = recalled
	switch m.Content.contentType {
	case MsgText:
		m.Content.Text = &MsgTextContent{
			Content: pb.Text.Content,
			Preview: pb.Text.Preview,
		}
	case MsgImage:
	}
}
