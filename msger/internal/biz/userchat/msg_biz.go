package userchat

import (
	"context"
	"encoding/json"

	"github.com/ryanreadbooks/whimer/msger/internal/infra"
	chatdao "github.com/ryanreadbooks/whimer/msger/internal/infra/dao/chat"
	"github.com/ryanreadbooks/whimer/msger/internal/model"

	"github.com/ryanreadbooks/whimer/misc/uuid"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xretry"
	"github.com/ryanreadbooks/whimer/misc/xsql"
)

type MsgBiz struct {
}

func NewMsgBiz() MsgBiz {
	return MsgBiz{}
}

type CreateMsgReq struct {
	Type    model.MsgType
	Content []byte
	Cid     string
	Ext     *MsgExt
}

// 创建一条消息
func (b *MsgBiz) CreateMsg(ctx context.Context, sender int64, req *CreateMsgReq) (*Msg, error) {
	var hasExt int8 = 0
	if req.Ext != nil {
		hasExt = 1
	}

	msgPo := &chatdao.MsgPO{
		Type:    req.Type,
		Status:  model.MsgStatusNormal,
		Sender:  sender,
		Content: req.Content, // TODO 需要加密
		Cid:     req.Cid,
		Ext:     hasExt,
	}

	err := xretry.OnError(func() error {
		msgId := uuid.NewUUID()
		msgPo.Id = msgId
		msgPo.Mtime = getAccurateTime()

		return infra.Dao().MsgDao.Create(ctx, msgPo)
	}, xsql.ErrDuplicate, 1) // retry on duplicate key error
	if err != nil {
		return nil, xerror.Wrapf(err, "msg dao create failed").WithCtx(ctx)
	}

	if hasExt > 0 {
		// 插入ext
		extPo := &chatdao.MsgExtPO{
			MsgId:     msgPo.Id,
			ImageKeys: makeJsonRawMessage(),
		}

		if req.Ext.ImageKeys != nil {
			extJson, err := json.Marshal(req.Ext.ImageKeys)
			if err != nil {
				return nil, xerror.Wrapf(err, "msg ext json marshal failed").WithCtx(ctx)
			}

			extPo.ImageKeys = extJson
		}

		err := infra.Dao().MsgExtDao.Create(ctx, extPo)
		if err != nil {
			return nil, xerror.Wrapf(err, "msg ext dao create failed").WithCtx(ctx)
		}
	}

	return &Msg{
		Id:     msgPo.Id,
		Type:   msgPo.Type,
		Status: msgPo.Status,
		Sender: msgPo.Sender,
		Mtime:  msgPo.Mtime,
		HasExt: msgPo.Ext > 0,
		Cid:    msgPo.Cid,
		Ext:    req.Ext,
	}, nil
}

// 消息绑定到会话中
func (b *MsgBiz) BindMsgToChat(ctx context.Context, msgId, chatId uuid.UUID, pos int64) error {
	chatMsgPo := &chatdao.ChatMsgPO{
		ChatId: chatId,
		MsgId:  msgId,
		Pos:    pos,
		Ctime:  getAccurateTime(),
	}

	err := infra.Dao().ChatMsgDao.Create(ctx, chatMsgPo)
	if err != nil {
		return xerror.Wrapf(err, "chat msg dao create failed").WithCtx(ctx)
	}

	return nil
}
