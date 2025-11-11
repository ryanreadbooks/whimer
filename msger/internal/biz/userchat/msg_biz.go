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

	// TODO content 解密

	msg, err := makeMsgFromPO(msgPo)
	if err != nil {
		return nil, err
	}

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

	return msg, nil
}

func (b *MsgBiz) BatchGetMsg(ctx context.Context, msgIds []uuid.UUID) (map[uuid.UUID]*Msg, error) {
	if len(msgIds) == 0 {
		return make(map[uuid.UUID]*Msg), nil
	}

	msgPoes, err := infra.Dao().MsgDao.BatchGetByIds(ctx, msgIds)
	if err != nil {
		return nil, xerror.Wrapf(err, "msg dao batch get failed").WithCtx(ctx)
	}

	// TODO content解密
	result := make(map[uuid.UUID]*Msg, len(msgPoes))
	hasExtIds := make([]uuid.UUID, 0, len(msgPoes))
	for _, chat := range msgPoes {
		msg, err := makeMsgFromPO(chat)
		if err != nil {
			continue
		}

		if msg.HasExt {
			hasExtIds = append(hasExtIds, msg.Id)
		}

		result[msg.Id] = msg
	}

	// fill exts
	extPoes, err := infra.Dao().MsgExtDao.BatchGetByIds(ctx, hasExtIds)
	if err != nil {
		return nil, xerror.Wrapf(err, "msg ext dao batch get failed").WithCtx(ctx)
	}

	for _, msg := range result {
		if msg.HasExt {
			if extPo, ok := extPoes[msg.Id]; ok {
				msg.Ext, _ = makeMsgExtFromPO(extPo)
			}
		}
	}

	return result, nil
}

// 创建一条消息
func (b *MsgBiz) CreateMsg(ctx context.Context, sender int64, req *CreateMsgReq) (*Msg, error) {
	msgPo := &chatdao.MsgPO{
		Type:    req.Type,
		Status:  model.MsgStatusNormal,
		Sender:  sender,
		Content: req.Content, // TODO 需要加密
		Cid:     req.Cid,
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

	return &Msg{
		Id:     msgPo.Id,
		Type:   msgPo.Type,
		Status: msgPo.Status,
		Sender: msgPo.Sender,
		Mtime:  msgPo.Mtime,
		HasExt: msgPo.Ext == hasMsgExt,
		Cid:    msgPo.Cid,
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

	msg, err := makeMsgFromPO(msgPo)
	if err != nil {
		return err
	}

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

func (b *MsgBiz) ListChatPos(ctx context.Context, chatId uuid.UUID, pos int64, count int32) ([]*ChatPos, error) {
	items, err := infra.Dao().ChatMsgDao.ListByPos(ctx, chatId, pos, count, true)
	if err != nil {
		return nil, xerror.Wrapf(err, "chat msg dao list by pos failed").
			WithCtx(ctx).WithExtras("chat_id", chatId, "pos", pos, "count", count)
	}

	result := make([]*ChatPos, 0, len(items))
	for _, r := range items {
		result = append(result, makeChatPosFromPO(r))
	}

	return result, nil
}
