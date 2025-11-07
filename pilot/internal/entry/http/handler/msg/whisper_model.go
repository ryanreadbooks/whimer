package msg

import (
	"github.com/ryanreadbooks/whimer/misc/xerror"
	whisper "github.com/ryanreadbooks/whimer/pilot/internal/biz/whisper/model"
	"github.com/ryanreadbooks/whimer/pilot/internal/model/errors"
)

type CreateWhisperChatReq struct {
	Target int64            `json:"target"`
	Type   whisper.ChatType `json:"type"`
}

func (r *CreateWhisperChatReq) Validate() error {
	if r == nil {
		return xerror.ErrNilArg
	}

	if r.Target == 0 {
		return errors.ErrUserNotFound
	}

	if ok := whisper.IsValidChatType(string(r.Type)); !ok {
		return errors.ErrUnsupportedChatType
	}

	return nil
}

type CreateWhisperChatResp struct {
	ChatId string `json:"chat_id"`
}

type SendWhisperChatReq struct{}

func (r *SendWhisperChatReq) Validate() error {

	return nil
}
