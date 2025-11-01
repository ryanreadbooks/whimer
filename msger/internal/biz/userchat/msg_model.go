package userchat

import (
	"encoding/json"

	"github.com/ryanreadbooks/whimer/misc/uuid"
	"github.com/ryanreadbooks/whimer/msger/internal/infra/dao/chat"
	"github.com/ryanreadbooks/whimer/msger/internal/model"
)

// 消息定义
type Msg struct {
	Id      uuid.UUID
	Type    model.MsgType
	Status  model.MsgStatus
	Sender  int64
	Mtime   int64
	Content []byte
	HasExt  bool
	Cid     string // 客户端侧id

	Ext *MsgExt // 消息扩展
}

func makeMsgFromPO(po *chat.MsgPO) *Msg {
	return &Msg{
		Id:      po.Id,
		Type:    po.Type,
		Status:  po.Status,
		Sender:  po.Sender,
		Mtime:   po.Mtime,
		Content: po.Content,
		HasExt:  po.Ext > 0,
		Cid:     po.Cid,
	}
}

type MsgExt struct {
	ImageKeys []string
}

func makeMsgExtFromPO(po *chat.MsgExtPO) (*MsgExt, error) {
	var ext MsgExt
	if err := json.Unmarshal(po.ImageKeys, &ext.ImageKeys); err != nil {
		return nil, err
	}

	return &ext, nil
}
