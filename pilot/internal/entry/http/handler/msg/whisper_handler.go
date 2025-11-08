package msg

import (
	"net/http"

	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
	whispermodel "github.com/ryanreadbooks/whimer/pilot/internal/biz/whisper/model"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 创建会话
func (h *Handler) CreateWhisperChat() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[CreateWhisperChatReq](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		var (
			ctx    = r.Context()
			uid    = metadata.Uid(ctx)
			chatId string
		)

		if req.Type == whispermodel.P2PChat {
			chatId, err = h.whisperBiz.CreateP2PChat(ctx, uid, req.Target)
		} else {
			// group chat
			chatId, err = h.whisperBiz.CreateGroupChat(ctx)
		}

		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, &CreateWhisperChatResp{ChatId: chatId})
	}
}

// 发消息
func (h *Handler) SendWhisperChatMsg() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[SendWhisperChatMsgReq](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		msgReq := &whispermodel.Msg{
			Type:    req.Type,
			Cid:     req.Cid,
			Content: req.Content,
		}
		msgReq.SetContentType()
		if err := msgReq.Validate(ctx); err != nil {
			xhttp.Error(r, w, err)
			return
		}

		// 往会话中发消息
		msgId, err := h.whisperBiz.SendChatMsg(ctx, req.ChatId, req.Cid, msgReq)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, &SendWhisperChatMsgResp{MsgId: msgId})
	}
}
