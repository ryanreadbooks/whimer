package msg

import (
	"net/http"

	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
	whisper "github.com/ryanreadbooks/whimer/pilot/internal/biz/whisper/model"
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

		if req.Type == whisper.P2PChat {
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
		req, err := xhttp.ParseValidate[SendWhisperChatReq](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		_ = req
	}
}
