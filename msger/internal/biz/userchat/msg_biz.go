package userchat

import (
	"context"
	"encoding/json"

	"github.com/ryanreadbooks/whimer/msger/internal/global"
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
	Type    model.MsgType `json:"type"`
	Content []byte        `json:"-"`
	Cid     string        `json:"cid"`
	Ext     *MsgExt       `json:"ext,omitempty"`
}

// 获取消息
func (b *MsgBiz) GetMsg(ctx context.Context, msgId uuid.UUID) (*Msg, error) {
	msgPo, err := infra.Dao().MsgDao.GetById(ctx, msgId)
	if err != nil {
		if xsql.IsNoRecord(err) {
			return nil, global.ErrMsgNotExist
		}

		return nil, xerror.Wrapf(err, "msg dao get by id failed").
			WithExtras("msg_id", msgId).
			WithCtx(ctx)
	}

	msg := makeMsgFromPO(msgPo)
	if msg.HasExt {
		// select ext
		msgExt, err := infra.Dao().MsgExtDao.GetById(ctx, msgId)
		if err != nil {
			return nil, xerror.Wrapf(err, "msg ext dao get failed").
				WithExtras("msg_id", msgId).
				WithCtx(ctx)
		}

		msg.Ext, err = makeMsgExtFromPO(msgExt)
		if err != nil {
			return nil, xerror.Wrapf(err, "msg ext json unmarshal failed").
				WithExtras("msg_id", msgId).
				WithCtx(ctx)
		}
	}

	// TODO msg.Content需要解密

	return msg, nil
}

// 创建一条消息
func (b *MsgBiz) CreateMsg(ctx context.Context, sender int64, req *CreateMsgReq) (*Msg, error) {
	var extFlag int8 = hasNoMsgExt
	if req.Ext != nil {
		extFlag = hasMsgExt
	}

	msgPo := &chatdao.MsgPO{
		Type:    req.Type,
		Status:  model.MsgStatusNormal,
		Sender:  sender,
		Content: req.Content, // TODO 需要加密
		Cid:     req.Cid,
		Ext:     extFlag,
	}

	err := xretry.OnError(func() error {
		msgId := uuid.NewUUID()
		msgPo.Id = msgId
		msgPo.Mtime = getAccurateTime()

		err := infra.Dao().MsgDao.Create(ctx, msgPo)
		return xerror.Wrapf(err, "msg dao create failed")
	}, xsql.ErrDuplicate, 1) // retry on duplicate key error
	// duplicate error基本都是cid重复
	if err != nil {
		return nil, xerror.Wrapf(err, "retry create msg failed").
			WithExtras("req", req).WithCtx(ctx)
	}

	if extFlag > 0 {
		// 插入ext
		extPo := &chatdao.MsgExtPO{
			MsgId:  msgPo.Id,
			Images: makeJsonRawMessage(),
		}

		if req.Ext.Images != nil {
			extJson, err := json.Marshal(req.Ext.Images)
			if err != nil {
				return nil, xerror.Wrapf(err, "msg ext json marshal failed").WithCtx(ctx)
			}

			extPo.Images = extJson
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
		HasExt: msgPo.Ext == hasMsgExt,
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

// 撤回消息
func (b *MsgBiz) RecallMsgById(ctx context.Context, uid int64, msgId uuid.UUID) error {
	msgPo, err := infra.Dao().MsgDao.GetById(ctx, msgId)
	if err != nil {
		if xsql.IsNoRecord(err) {
			return xerror.Wrap(global.ErrMsgNotExist)
		}

		return xerror.Wrapf(err, "msg dao get by id failed").WithCtx(ctx).WithExtras("msg_id", msgId)
	}

	msg := makeMsgFromPO(msgPo)

	return b.recallMsg(ctx, uid, msg)
}

func (b *MsgBiz) RecallMsg(ctx context.Context, uid int64, msg *Msg) error {
	return b.recallMsg(ctx, uid, msg)
}

func (b *MsgBiz) recallMsg(ctx context.Context, uid int64, msg *Msg) error {
	if msg.IsStatusRecalled() {
		return global.ErrMsgAlreadyRecalled
	}

	msgId := msg.Id
	mtime := getAccurateTime()
	// first set new status for msg
	err := infra.Dao().MsgDao.UpdateStatus(ctx, msgId, model.MsgStatusRecall, mtime)
	if err != nil {
		return xerror.Wrapf(err, "msg dao update status failed").
			WithExtras("msg_id", msgId).WithCtx(ctx)
	}

	recallExt := &MsgRecall{
		Uid:  uid,
		Time: mtime,
	}

	recallData, err := json.Marshal(recallExt)
	if err != nil {
		return xerror.Wrapf(err, "json marshal recall ext failed").
			WithExtras("ext", recallExt).WithCtx(ctx)
	}

	// then set ext status
	err = infra.Dao().MsgExtDao.SetRecall(ctx, msgId, recallData)
	if err != nil {
		return xerror.Wrapf(err, "msg ext dao set recall failed").
			WithExtras("msg_id", msgId).
			WithCtx(ctx)
	}

	return nil
}

func (b *MsgBiz) BatchGetMsgPos(ctx context.Context, 
	chatId uuid.UUID, msgIds []uuid.UUID) (map[uuid.UUID]int64, error) {

	msgPos, err := infra.Dao().ChatMsgDao.BatchGetPos(ctx, chatId, msgIds)
	if err != nil {
		return nil, xerror.Wrapf(err, "chat msg dao batch get pos failed").WithCtx(ctx)
	}

	return msgPos, nil
}
