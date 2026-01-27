package msg

import (
	"net/http"

	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
	"github.com/ryanreadbooks/whimer/pilot/internal/app/whisper/dto"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func (h *Handler) CreateWhisperChat() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cmd, err := xhttp.ParseValidateJsonBody[dto.CreateP2PChatCommand](r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		var (
			ctx    = r.Context()
			chatId string
		)
		cmd.Uid = metadata.Uid(ctx)

		if cmd.IsP2P() {
			chatId, err = h.whisperApp.CreateP2PChat(ctx, cmd)
		} else {
			chatId, err = h.whisperApp.CreateGroupChat(ctx)
		}

		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, &dto.CreateChatResult{ChatId: chatId})
	}
}

func (h *Handler) SendWhisperChatMsg() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cmd, err := xhttp.ParseValidateJsonBody[dto.SendChatMsgCommand](r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		msgId, err := h.whisperApp.SendChatMsg(r.Context(), cmd)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, &dto.SendChatMsgResult{MsgId: msgId})
	}
}

func (h *Handler) RecallWhisperChatMsg() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cmd, err := xhttp.ParseValidateJsonBody[dto.RecallChatMsgCommand](r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		if err = h.whisperApp.RecallChatMsg(r.Context(), cmd); err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, nil)
	}
}

func (h *Handler) ListWhisperRecentChats() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query, err := xhttp.ParseValidate[dto.ListRecentChatsQuery](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		query.Uid = metadata.Uid(r.Context())

		result, err := h.whisperApp.ListRecentChats(r.Context(), query)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, result)
	}
}

func (h *Handler) ListWhisperChatMsgs() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query, err := xhttp.ParseValidate[dto.ListChatMsgsQuery](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		query.Uid = metadata.Uid(r.Context())

		msgs, err := h.whisperApp.ListChatMsgs(r.Context(), query)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, msgs)
	}
}

func (h *Handler) ClearWhisperChatUnread() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cmd, err := xhttp.ParseValidateJsonBody[dto.ClearChatUnreadCommand](r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		if err = h.whisperApp.ClearChatUnread(r.Context(), cmd); err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, nil)
	}
}
