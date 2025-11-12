package msg

import (
	"github.com/ryanreadbooks/whimer/misc/xerror"
	whispermodel "github.com/ryanreadbooks/whimer/pilot/internal/biz/whisper/model"
	"github.com/ryanreadbooks/whimer/pilot/internal/model/errors"
)

type CreateWhisperChatReq struct {
	Target int64                 `json:"target"`
	Type   whispermodel.ChatType `json:"type"`
}

func (r *CreateWhisperChatReq) Validate() error {
	if r == nil {
		return xerror.ErrNilArg
	}

	if r.Target == 0 {
		return errors.ErrUserNotFound
	}

	if ok := whispermodel.IsValidChatType(string(r.Type)); !ok {
		return errors.ErrUnsupportedChatType
	}

	return nil
}

type CreateWhisperChatResp struct {
	ChatId string `json:"chat_id"`
}

type SendWhisperChatMsgReq struct {
	ChatId  string                   `json:"chat_id"`
	Type    whispermodel.MsgType     `json:"type"`
	Cid     string                   `json:"cid"`
	Content *whispermodel.MsgContent `json:"content"`
}

func (r *SendWhisperChatMsgReq) Validate() error {
	if r == nil {
		return xerror.ErrNilArg
	}

	// 详细校验在biz中进行
	return nil
}

type SendWhisperChatMsgResp struct {
	MsgId string `json:"msg_id"`
}

type ListWhisperRecentChatsReq struct {
	Cursor string `form:"cursor,optional"`
	Count  int32  `form:"count,default=30"`
}

type ListWhisperRecentChatsResp struct {
	Items      []*whispermodel.RecentChat `json:"items"`
	HasNext    bool                       `json:"has_next"`
	NextCursor string                     `json:"next_cursor"`
}

type ListWhisperChatMsgsReq struct {
	ChatId string `form:"chat_id"`
	Pos    int64  `form:"pos,optional"`
	Count  int32  `form:"count,default=50"`
}
