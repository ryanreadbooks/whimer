package msg

import (
	"context"
	"net/http"

	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
	"github.com/ryanreadbooks/whimer/misc/xslice"
	"github.com/ryanreadbooks/whimer/pilot/internal/biz/sysnotify/model"
	usermodel "github.com/ryanreadbooks/whimer/pilot/internal/biz/user/model"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func (h *Handler) ListSysMsgMentions() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[CursorAndCountReq](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		var (
			ctx = r.Context()
			uid = metadata.Uid(ctx)
		)

		result, err := h.sysNotifyBiz.ListUserMentionMsg(ctx, uid, req.Cursor, req.Count)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		output, err := attachUserToSysMsgs(ctx, h, result.ChatId, result.HasNext, result.Msgs,
			func(msg *model.MentionedMsg) int64 { return msg.Uid },
			func(msg *model.MentionedMsg, user *usermodel.User) *SystemMsgForMention {
				return &SystemMsgForMention{
					MentionedMsg: msg,
					User:         user,
				}
			},
		)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, output)
	}
}

func (h *Handler) ListSysMsgReplies() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[CursorAndCountReq](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		var (
			ctx = r.Context()
			uid = metadata.Uid(ctx)
		)

		result, err := h.sysNotifyBiz.ListUserReplyMsg(ctx, uid, req.Cursor, req.Count)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		output, err := attachUserToSysMsgs(ctx, h, result.ChatId, result.HasNext, result.Msgs,
			func(msg *model.ReplyMsg) int64 { return msg.Uid },
			func(msg *model.ReplyMsg, user *usermodel.User) *SystemMsgForReply {
				return &SystemMsgForReply{
					ReplyMsg: msg,
					User:     user,
				}
			},
		)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, output)
	}
}

func (h *Handler) ListSysMsgLikes() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[CursorAndCountReq](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		var (
			ctx = r.Context()
			uid = metadata.Uid(ctx)
		)

		result, err := h.sysNotifyBiz.ListUserLikeMsg(ctx, uid, req.Cursor)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		output, err := attachUserToSysMsgs(ctx, h, result.ChatId, result.HasNext, result.Msgs,
			func(msg *model.LikesMsg) int64 {
				return msg.Uid
			}, func(msg *model.LikesMsg, u *usermodel.User) *SystemMsgForLikes {
				return &SystemMsgForLikes{
					LikesMsg: msg,
					User:     u,
				}
			})
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, output)
	}
}

func attachUserToSysMsgs[T any, R any](ctx context.Context,
	h *Handler,
	chatId string, hasNext bool,
	msgs []*T,
	getUidFn func(*T) int64,
	genItem func(*T, *usermodel.User) R,
) (*ListSysMsgResp[R], error) {
	uids := make([]int64, 0, len(msgs))
	for _, msg := range msgs {
		uids = append(uids, getUidFn(msg))
	}

	uids = xslice.Uniq(uids)

	users, err := h.userBiz.ListUsersV2(ctx, uids)
	if err != nil {
		return nil, err
	}

	resultMsgs := make([]R, 0, len(msgs))
	for _, msg := range msgs {
		user := users[getUidFn(msg)]
		resultMsgs = append(resultMsgs, genItem(msg, user))
	}

	return &ListSysMsgResp[R]{
		ChatId:  chatId,
		HasNext: hasNext,
		Msgs:    resultMsgs,
	}, nil
}
