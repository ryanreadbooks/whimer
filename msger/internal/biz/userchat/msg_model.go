package userchat

import (
	"encoding/json"

	"github.com/ryanreadbooks/whimer/misc/uuid"
	"github.com/ryanreadbooks/whimer/msger/internal/infra/dao/chat"
	"github.com/ryanreadbooks/whimer/msger/internal/model"
)

const (
	hasNoMsgExt int8 = 0
	hasMsgExt   int8 = 1
)

// 消息定义
type Msg struct {
	Id      uuid.UUID
	Type    model.MsgType
	Status  model.MsgStatus
	Sender  int64
	Mtime   int64
	Content MsgContent
	HasExt  bool
	Cid     string // 客户端侧id

	Ext *MsgExt // 消息扩展
}

func (m *Msg) IsStatusRecalled() bool {
	return m != nil && m.Status == model.MsgStatusRecall
}

func (m *Msg) IsStatusNormal() bool {
	return m != nil && m.Status == model.MsgStatusNormal
}

func makeMsgFromPO(po *chat.MsgPO) (*Msg, error) {
	ct, _, err := ParseMsgContent(po.Content)
	if err != nil {
		return nil, err
	}

	return &Msg{
		Id:      po.Id,
		Type:    po.Type,
		Status:  po.Status,
		Sender:  po.Sender,
		Mtime:   po.Mtime,
		Content: ct,
		HasExt:  po.Ext == hasMsgExt,
		Cid:     po.Cid,
	}, nil
}

type MsgExt struct {
	Recall *MsgRecall
}

type MsgRecall struct {
	Uid  int64 `json:"uid"`  // 撤回消息人
	Time int64 `json:"time"` // 撤回时间
}

func (r *MsgRecall) GetUid() int64 {
	if r != nil {
		return r.Uid
	}

	return 0
}

func (r *MsgRecall) GetTime() int64 {
	if r != nil {
		return r.Time
	}

	return 0
}

func makeMsgExtFromPO(po *chat.MsgExtPO) (*MsgExt, error) {
	var ext MsgExt
	if po == nil {
		return &ext, nil
	}

	if err := json.Unmarshal(po.Recall, &ext.Recall); err != nil {
		return nil, err
	}

	return &ext, nil
}
