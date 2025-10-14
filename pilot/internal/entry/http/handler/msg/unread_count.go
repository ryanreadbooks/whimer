package msg

import (
	"net/http"

	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
	sysmsgmodel "github.com/ryanreadbooks/whimer/pilot/internal/biz/sysnotify/model"
	"github.com/zeromicro/go-zero/rest/httpx"
)

type GetTotalUnreadCountResp struct {
	System     *sysmsgmodel.ChatsUnreadCount `json:"system"`
	PeerUnread int64                         `json:"peer_unread"`
}

func (h *Handler) GetTotalUnreadCount() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			ctx = r.Context()
			uid = metadata.Uid(ctx)
		)

		// 1. 系统会话未读
		sysUnreads, err := h.sysNotifyBiz.GetChatUnread(ctx, uid)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, &GetTotalUnreadCountResp{System: sysUnreads})
	}
}

// 清除未读数
func (h *Handler) ClearChatUnread() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[SysChatReq](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		var (
			ctx = r.Context()
			uid = metadata.Uid(ctx)
		)

		err = h.sysNotifyBiz.ClearChatUnread(ctx, uid, req.ChatId)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, nil)
	}
}
