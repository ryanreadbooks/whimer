package msg

import (
	"net/http"

	msgv1 "github.com/ryanreadbooks/whimer/msger/api/p2p/v1"

	"github.com/ryanreadbooks/whimer/api-x/internal/config"
	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
	"github.com/zeromicro/go-zero/rest/httpx"
)

type Handler struct{}

func NewHandler(c *config.Config) *Handler {
	return &Handler{}
}

func (h *Handler) ListChats() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[ListChatsReq](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		if req.Seq < 0 {
			req.Seq = 0
		}
		if req.Count <= 0 || req.Count > 20 {
			req.Count = 20
		}

		ctx := r.Context()
		uid := metadata.Uid(ctx)
		resp, err := chatter.ListChat(ctx, &msgv1.ListChatRequest{
			UserId: uid,
			Seq:    req.Seq,
			Count:  int32(req.Count),
		})
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, resp)
	}
}
