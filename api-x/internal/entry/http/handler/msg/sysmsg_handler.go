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

		mentionMsgs, hasNext, err := h.sysNotifyMsgBiz.ListUserMentionMsg(ctx, uid, req.Cursor, req.Count)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		// filter users
		sourceUids := make([]int64, 0, len(mentionMsgs))
		resultMsgs := make([]*SystemMsgForMention, 0, len(mentionMsgs))
		uidMsgMappings := make(map[int64][]*SystemMsgForMention, len(mentionMsgs))

		for _, msg := range mentionMsgs {
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

		xhttp.OkJson(w, &ListSysMsgMentionsResp{Msgs: resultMsgs, HasNext: hasNext})
	}
}
