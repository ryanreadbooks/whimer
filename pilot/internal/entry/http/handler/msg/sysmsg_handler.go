package msg

import (
	"net/http"

	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
	"github.com/ryanreadbooks/whimer/misc/xslice"
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

		// filter users
		mLen := len(result.Msgs)
		sourceUids := make([]int64, 0, mLen)
		resultMsgs := make([]*SystemMsgForMention, 0, mLen)
		uidMsgMappings := make(map[int64][]*SystemMsgForMention, mLen)

		for _, msg := range result.Msgs {
			sourceUids = append(sourceUids, msg.Uid)
			tmp := &SystemMsgForMention{
				MentionedMsg: msg,
			}
			resultMsgs = append(resultMsgs, tmp)
			uidMsgMappings[msg.Uid] = append(uidMsgMappings[msg.Uid], tmp)
		}

		// find users
		sourceUids = xslice.Uniq(sourceUids)
		sourceUsers, err := h.userBiz.ListUsersV2(ctx, sourceUids)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		for _, resultMsg := range resultMsgs {
			resultMsg.User = sourceUsers[resultMsg.MentionedMsg.Uid]
		}

		xhttp.OkJson(w, &ListSysMsgMentionsResp{
			ChatId:  result.ChatId,
			Msgs:    resultMsgs,
			HasNext: result.HasNext,
		})
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

		var (
			mLen           = len(result.Msgs)
			srcUids        = make([]int64, 0, mLen)
			resultMsgs     = make([]*SystemMsgForReply, 0, mLen)
			uidMsgMappings = make(map[int64][]*SystemMsgForReply, mLen)
		)
		// fill source user
		for _, msg := range result.Msgs {
			srcUids = append(srcUids, msg.Uid)
			tmp := &SystemMsgForReply{
				ReplyMsg: msg,
			}
			resultMsgs = append(resultMsgs, tmp)
			uidMsgMappings[msg.Uid] = append(uidMsgMappings[msg.Uid], tmp)
		}

		// find users
		srcUids = xslice.Uniq(srcUids)
		sourceUsers, err := h.userBiz.ListUsersV2(ctx, srcUids)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		// assign users
		for _, resultMsg := range resultMsgs {
			resultMsg.User = sourceUsers[resultMsg.ReplyMsg.Uid]
		}

		xhttp.OkJson(w, &ListSysMsgRepliesResp{
			ChatId:  result.ChatId,
			HasNext: result.HasNext,
			Msgs:    resultMsgs,
		})
	}
}
