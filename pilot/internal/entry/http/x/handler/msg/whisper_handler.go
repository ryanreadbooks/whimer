package msg

import (
	"net/http"

	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/misc/xslice"
	whispermodel "github.com/ryanreadbooks/whimer/pilot/internal/biz/whisper/model"
)

// 创建会话
func (h *Handler) CreateWhisperChat() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidateJsonBody[CreateWhisperChatReq](r)
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
		req, err := xhttp.ParseValidateJsonBody[SendWhisperChatMsgReq](r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		msgReq := &whispermodel.MsgReq{
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

// 撤回消息
func (h *Handler) RecallWhisperChatMsg() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidateJsonBody[RecallWhisperChatMsgReq](r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		err = h.whisperBiz.RecallChatMsg(ctx, req.ChatId, req.MsgId)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, nil)
	}
}

func (h *Handler) ListWhisperRecentChats() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidateForm[ListWhisperRecentChatsReq](r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		var (
			ctx = r.Context()
			uid = metadata.Uid(ctx)
		)

		recentChats, pageResult, err := h.whisperBiz.ListRecentChats(ctx, uid, req.Cursor, req.Count)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, &ListWhisperRecentChatsResp{
			Items:      recentChats,
			HasNext:    pageResult.HasNext,
			NextCursor: pageResult.NextCursor,
		})
	}
}

func (h *Handler) ListWhisperChatMsgs() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidateForm[ListWhisperChatMsgsReq](r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		var (
			ctx = r.Context()
			uid = metadata.Uid(ctx)
		)

		msgs, err := h.whisperBiz.ListChatMsgs(ctx, uid, req.ChatId, req.Pos, req.Count)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		// collect uids
		senderUids := xslice.Extract(msgs, func(t *whispermodel.Msg) int64 {
			return t.SenderUid
		})
		senderUids = xslice.Uniq(senderUids)

		userInfos, err := h.userAdapter.BatchGetUser(ctx, senderUids)
		if err != nil {
			xlog.Msgf("user adapter batch get user failed").Err(err).Errorx(ctx)
		} else {
			for _, msg := range msgs {
				if user, ok := userInfos[msg.SenderUid]; ok {
					msg.SetSenderFromVo(user)
				}
			}
		}

		xhttp.OkJson(w, msgs)
	}
}

func (h *Handler) ClearWhisperChatUnread() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidateJsonBody[ChatIdParamReq](r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		err = h.whisperBiz.ClearChatUnread(ctx, req.ChatId)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, nil)
	}
}
