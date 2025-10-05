package msg

import (
	"net/http"

	"github.com/ryanreadbooks/whimer/msger/api/msg"
	msgv1 "github.com/ryanreadbooks/whimer/msger/api/p2p/v1"

	"github.com/ryanreadbooks/whimer/api-x/internal/config"
	"github.com/ryanreadbooks/whimer/api-x/internal/infra"
	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
	"github.com/zeromicro/go-zero/rest/httpx"
)

type Handler struct{}

func NewHandler(c *config.Config) *Handler {
	return &Handler{}
}

// 发起会话
func (h *Handler) CreateChat() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[CreateChatReq](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()

		resp, err := infra.Chatter().CreateChat(ctx, &msgv1.CreateChatRequest{
			Initiator: metadata.Uid(ctx),
			Target:    req.Target,
		})
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, resp)
	}
}

// 拉会话列表
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
		resp, err := infra.Chatter().ListChat(ctx, &msgv1.ListChatRequest{
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

type GetChatReq struct {
	Id int64 `form:"id"`
}

func (r *GetChatReq) Validate() error {
	if r.Id == 0 {
		return xerror.ErrArgs.Msg("会话不存在")
	}

	return nil
}

func (h *Handler) GetChat() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[GetChatReq](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}
		ctx := r.Context()
		uid := metadata.Uid(ctx)

		resp, err := infra.Chatter().GetChat(ctx, &msgv1.GetChatRequest{
			UserId: uid,
			ChatId: req.Id,
		})
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, resp)
	}
}

// 拉消息列表
func (h *Handler) ListMsgs() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[ListMsgsReq](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		uid := metadata.Uid(ctx)
		messages, err := infra.Chatter().ListMsg(ctx, &msgv1.ListMsgRequest{
			ChatId: req.ChatId,
			UserId: uid,
			Seq:    req.Seq,
			Count:  int32(req.Count),
		})
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, messages)
	}
}

// 发消息
func (h *Handler) SendMsg() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[SendMsgReq](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		sender := metadata.Uid(ctx)

		// TODO check if sender can send msg to receiver

		resp, err := infra.Chatter().SendMsg(ctx, &msgv1.SendMsgRequest{
			Sender:   sender,
			Receiver: req.Receiver,
			ChatId:   req.ChatId,
			Msg: &msg.MsgContent{
				Type: msg.MsgType(req.MsgType),
				Data: req.Content,
			},
		})
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, resp)
	}
}

// 删除会话
func (h *Handler) DeleteChat() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[DeleteChatReq](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		uid := metadata.Uid(ctx)

		// TODO check uid can delete chat or not

		_, err = infra.Chatter().DeleteChat(ctx, &msgv1.DeleteChatRequest{
			UserId: uid,
			ChatId: req.ChatId,
		})
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, nil)
	}
}

// 删除消息
func (h *Handler) DeleteMsg() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[DeleteMsgReq](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		uid := metadata.Uid(ctx)

		// TODO check uid can delete message or not

		_, err = infra.Chatter().DeleteMsg(ctx, &msgv1.DeleteMsgRequest{
			UserId: uid,
			ChatId: req.ChatId,
			MsgId:  req.MsgId,
		})
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, nil)
	}
}
